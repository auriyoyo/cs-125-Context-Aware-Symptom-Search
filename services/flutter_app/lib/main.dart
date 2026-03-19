import 'dart:async';
import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:geolocator/geolocator.dart';

void main() {
  runApp(const MyApp());
}

class LocationService {
  Future<Map<String, String>> getLocationData() async {
    final pos = await Geolocator.getCurrentPosition();
    return await _getAddressFromCoordinatesWeb(pos.latitude, pos.longitude);
  }

  Future<Map<String, String>> _getAddressFromCoordinatesWeb(
        double lat, double lng) async {
      final apiKey = "AIzaSyAFfhaJa7Onm30sSTRnWeBmV5r_ALcZrUY"; // replace with your key
      final url =
          "https://maps.googleapis.com/maps/api/geocode/json?latlng=$lat,$lng&key=$apiKey";
      final res = await http.get(Uri.parse(url));
      final data = jsonDecode(res.body);
      if (data['status'] == 'OK' && data['results'].isNotEmpty) {
        final components = data['results'][0]['address_components'];
        String city = '';
        String state = '';
        String zip = '';
        for (var c in components) {
          if (c['types'].contains('locality')) city = c['long_name'];
          if (c['types'].contains('administrative_area_level_1')) state = c['short_name'];
          if (c['types'].contains('postal_code')) zip = c['long_name'];
        }
        return {"city": city, "state": state, "zipCode": zip, "country": "US"};
      }
      return {"city": "", "state": "", "zipCode": "", "country": "US"};
    }
}

class ApiService {
  Future<void> sendLocation(Map<String, dynamic> payload) async {
    final response = await http.post(
      Uri.parse('http://localhost:8080/events/location'),
      headers: {"Content-Type": "application/json"},
      body: jsonEncode(payload),
    );

    if (response.statusCode != 201) {
      throw Exception("Failed to send location");
    }
  }
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Context Aware Symptom Search',
      theme: ThemeData(
        colorScheme: .fromSeed(seedColor: const Color.fromARGB(255, 58, 169, 183)),
      ),
      home: const MyHomePage(title: 'Context Aware Symptom Search'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});
  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  String _disease = "Welcome to the Context Aware Symptom Search! Enter a symptom to begin searching for diseases!";
  String _live = "Please enable location services in order to get live alerts!";
  final List<TextEditingController> _controllers = [];
  final List<bool> _isEditable = [];
  final locationService = LocationService();
  final apiService = ApiService();

  //Search
  Future<void> fetchDisease(String query) async {
    final structure = "http://localhost:8080/search?user=thomas&symptoms=";
    final url = Uri.parse(structure + query);
    final response = await http.get(url);

    setState(() {
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        
        final List results = data['results'];
        if (results.isEmpty) {
          _disease = "No diseases matched your symptoms.";
          return;
        }

        String formatted = "";
        for (var item in results) {
          final disease = item['result']['disease'];
          final score = item['result']['score'];
          final reasons = item['reasons']; // List<String>

          formatted += 'Score: $score - "$disease"\n';
          for (var r in reasons) {
            formatted += '  $r\n';
          }
          formatted += '\n'; // space between diseases
        }

        _disease = formatted;
      } else {
        _disease = 'Enter a symptom to begin searching for diseases!';
      }
    });
  }

  Future<void> fetchLive() async {
    final location = await http.get(Uri.parse("http://localhost:8080/context/user/events?user=test_user"));
    //final structure = "http://localhost:8080/context/area/risks?zip=";
    final locationList = jsonDecode(location.body) as List;
    if (locationList.isEmpty) return;

    final locationData = locationList[0];
    final zip = locationData['zip_code'];

    final liveResponse = await http.get(Uri.parse("http://localhost:8080/context/area/risks?zip=$zip"));
    final List live = jsonDecode(liveResponse.body)['active_risks'];

    setState(() {
      if (live.isEmpty) {
        _live = "No active risks found in your location! [Zip Code: $zip]";
        return;
      } else {
        String formatted = "";
        for (var item in live) {
          final hazard = item['hazard'].toUpperCase();
          final severity = item['severity'].toUpperCase();
          final List symptoms = item['symptomTags'];

          formatted += '$hazard of $severity severity in your area is causing: ';
          for (var s in symptoms) {
            s = s.toUpperCase();
            formatted += '$s ';
          }
          formatted += '\n\n';
        }

        _live = formatted;
        return;
      }
    });
  }

  Future<void> updateLocation() async {
    final location = await locationService.getLocationData();

    final payload = {
      "userId": "test_user",
      ...location,
    };

    await apiService.sendLocation(payload);
  }


  @override
  void initState() {
    super.initState();
    updateLocation();
    fetchLive();
    _addNewField(); // Start with one field
  }

  void _addNewField() {
    _controllers.add(TextEditingController());
    _isEditable.add(true);
  }

  void _query() {
    String query = _controllers
      .map((c) => c.text.trim())
      .where((text) => text.isNotEmpty)
      .join(',');
    fetchDisease(query);
  }

  @override
  void dispose() {
    for (var controller in _controllers) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(widget.title),
        centerTitle: true,
      ),
      body: Column(
        children: [
          // Main scrollable content
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(_disease),
                const SizedBox(height: 20),

                ...List.generate(_controllers.length, (index) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _controllers[index],
                            readOnly: !_isEditable[index],
                            onSubmitted: (value) {
                              _query();
                              setState(() {
                                _isEditable[index] = false;
                                _addNewField();
                              });
                            },
                            decoration: const InputDecoration(
                              border: OutlineInputBorder(),
                              hintText: 'Enter symptom',
                            ),
                          ),
                        ),

                        if (!_isEditable[index])
                          IconButton(
                            icon: const Icon(Icons.close, color: Colors.red),
                            onPressed: () {
                              setState(() {
                                _controllers[index].dispose();
                                _controllers.removeAt(index);
                                _isEditable.removeAt(index);
                              });
                              _query();
                            },
                          ),
                      ],
                    ),
                  );
                }),
              ],
            ),
          ),

          // Divider line
          const Divider(height: 1, color: Colors.black54),

          // Live alerts box at the bottom
          Container(
            color: Colors.grey[200],
            padding: const EdgeInsets.all(16),
            width: double.infinity,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  "Live Alerts",
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                ),
                const SizedBox(height: 8),
                Text(
                  _live,
                  style: const TextStyle(fontSize: 14),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
