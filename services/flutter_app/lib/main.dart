import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;

void main() {
  runApp(const MyApp());
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
  final List<TextEditingController> _controllers = [];
  final List<bool> _isEditable = [];

  //Search
  Future<void> fetchDisease(String query) async {
    final structure = "http://localhost:8080/search?user=thomas&symptoms=";
    final url = Uri.parse(structure + query);
    final response = await http.get(url);

    setState(() {
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        
        _disease = data.toString();
      } else {
        _disease = 'Request failed with status: ${response.statusCode}';
      }
    });
  }

  @override
  void initState() {
    super.initState();
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
    // This method is rerun every time setState is called
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(widget.title),
        centerTitle: true,
      ),
      body: ListView(
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
    );

  }
}
