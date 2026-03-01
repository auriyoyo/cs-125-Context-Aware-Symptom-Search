import os
from typing import List, Dict, Any

from pymongo import MongoClient, UpdateOne, ASCENDING
from pymongo.database import Database
from pymongo.collection import Collection


def _client_options() -> dict:
    opts = {}
    if os.getenv("MONGO_TLS_INSECURE", "").lower() in ("1", "true", "yes"):
        opts["tlsAllowInvalidCertificates"] = True
    return opts


class KaggleDatasetDB:
    def __init__(self, uri: str, db_name: str = "kaggle_dataset"):
        self._client = MongoClient(uri, **_client_options())
        self._db: Database = self._client[db_name]
        self._diseases: Collection = self._db["diseases"]
        self._symptoms: Collection = self._db["symptoms"]

    def replace_by_source(self, collection: str, docs: List[Dict[str, Any]], source_tag: str) -> None:
        coll = self._diseases if collection == "diseases" else self._symptoms
        coll.delete_many({"source": source_tag})
        if docs:
            ops = [UpdateOne({"_id": d["_id"]}, {"$set": d}, upsert=True) for d in docs]
            res = coll.bulk_write(ops, ordered=False)
            print(f"{collection} bulk_write: {res.bulk_api_result}")

    def ensure_indexes(self) -> None:
        self._diseases.create_index([("symptoms", ASCENDING)])
        self._diseases.create_index([("name", ASCENDING)])
        self._diseases.create_index([("source", ASCENDING)])
        self._symptoms.create_index([("diseases", ASCENDING)])
        self._symptoms.create_index([("name", ASCENDING)])
        self._symptoms.create_index([("source", ASCENDING)])

    def close(self) -> None:
        self._client.close()
