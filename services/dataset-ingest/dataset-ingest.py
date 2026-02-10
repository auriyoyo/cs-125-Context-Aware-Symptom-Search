import os
from collections import defaultdict
from typing import Dict, Any, List, Tuple

import pandas as pd
from pymongo import MongoClient, UpdateOne, ASCENDING
from dotenv import load_dotenv


SOURCE_TAG = "kaggle:diseases-and-symptoms-dataset:onehot"


def repo_root() -> str:
    # dataset-ingest/dataset-ingest.py
    return os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


def main():
    # load .env from repo root
    load_dotenv(os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".env")))

    mongo_uri = os.getenv("MONGODB_URI") or os.getenv("MONGO_URI")
    mongo_db = os.getenv("MONGO_DB", "symptomsearch")
    coll_name = os.getenv("MONGO_DISEASES_COLLECTION", "diseases")

    print("Loaded .env from:", os.path.join(repo_root(), ".env"))
    print("MONGODB_URI is:", "SET" if os.getenv("MONGODB_URI") else "MISSING")

    csv_path = os.getenv("DATASET_CSV_PATH")
    if not csv_path:
        # default location
        csv_path = os.path.join(repo_root(), "dataset-ingest", "data",
                                "Final_Augmented_dataset_Diseases_and_Symptoms.csv")
    else:
        # interpret relative to repo root
        if not os.path.isabs(csv_path):
            csv_path = os.path.join(repo_root(), csv_path)

    print("Mongo DB:", mongo_db)
    print("Collection:", coll_name)
    print("CSV path:", csv_path)

    if not mongo_uri:
        raise ValueError("Missing MONGODB_URI in .env")

    if not os.path.exists(csv_path):
        raise FileNotFoundError(f"CSV not found at: {csv_path}")

    print("Reading CSV (big file)...")
    df = pd.read_csv(csv_path)

    if "diseases" not in df.columns:
        raise ValueError(f"Expected a 'diseases' column. Found: {list(df.columns)[:20]} ...")

    disease_col = "diseases"
    symptom_cols = [c for c in df.columns if c != disease_col]

    print(f"Rows: {len(df):,}")
    print(f"Unique diseases: {df[disease_col].nunique():,}")
    print(f"Symptom columns: {len(symptom_cols):,}")

    # disease -> symptom -> count
    symptom_counts: Dict[str, Dict[str, int]] = defaultdict(lambda: defaultdict(int))
    disease_row_counts: Dict[str, int] = defaultdict(int)

    # iterate rows
    for row in df.itertuples(index=False):
        disease = getattr(row, disease_col)
        disease = str(disease).strip().lower()
        if not disease:
            continue

        disease_row_counts[disease] += 1

        for col_name, val in zip(df.columns, row):
            if col_name == disease_col:
                continue
            try:
                if int(val) == 1:
                    symptom_counts[disease][col_name.strip().lower()] += 1
            except Exception:
                continue

    # build one doc per disease
    docs: List[Dict[str, Any]] = []
    for disease, counts in symptom_counts.items():
        sorted_symptoms: List[Tuple[str, int]] = sorted(counts.items(), key=lambda x: x[1], reverse=True)
        docs.append({
            "_id": disease,
            "name": disease,
            "source": SOURCE_TAG,
            "n_samples": disease_row_counts[disease],
            "symptoms": [s for s, _ in sorted_symptoms],
            "symptom_counts": [{"symptom": s, "count": c} for s, c in sorted_symptoms],
        })

    print(f"Prepared {len(docs):,} disease docs. Connecting to Mongo...")

    client = MongoClient(mongo_uri)
    db = client[mongo_db]
    coll = db[coll_name]

    # remove previous import from this source
    coll.delete_many({"source": SOURCE_TAG})

    ops = [
        UpdateOne({"_id": d["_id"]}, {"$set": d}, upsert=True)
        for d in docs
    ]
    if ops:
        res = coll.bulk_write(ops, ordered=False)
        print("Mongo bulk_write:", res.bulk_api_result)

    # helpful indexes
    coll.create_index([("name", ASCENDING)])
    coll.create_index([("symptoms", ASCENDING)])
    coll.create_index([("source", ASCENDING)])

    print("DONE")


if __name__ == "__main__":
    main()
