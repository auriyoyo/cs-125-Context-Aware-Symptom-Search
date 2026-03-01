import os
import glob
from collections import defaultdict
from typing import Dict, Any, List, Tuple

import pandas as pd
from dotenv import load_dotenv
import kagglehub

from mongo_client import KaggleDatasetDB

SOURCE_TAG = "kaggle:diseases-and-symptoms-dataset:onehot"


def repo_root() -> str:
    return os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))


def resolve_csv_path() -> str:
    load_dotenv(os.path.join(repo_root(), ".env"))
    csv_path = os.getenv("DATASET_CSV_PATH")
    if csv_path:
        if not os.path.isabs(csv_path):
            csv_path = os.path.join(repo_root(), csv_path)
        if not os.path.exists(csv_path):
            raise FileNotFoundError(f"CSV not found at: {csv_path}")
        return csv_path
    path = kagglehub.dataset_download("dhivyeshrk/diseases-and-symptoms-dataset")
    candidates = glob.glob(os.path.join(path, "**/*.csv"), recursive=True)
    if not candidates:
        raise FileNotFoundError(f"No CSV found under {path}")
    preferred = [c for c in candidates if "Final_Augmented" in c or "Diseases_and_Symptoms" in c]
    return preferred[0] if preferred else candidates[0]


def parse_csv(csv_path: str) -> Tuple[Dict[str, Dict[str, int]], Dict[str, int], Dict[str, set]]:
    df = pd.read_csv(csv_path)
    if "diseases" not in df.columns:
        raise ValueError(f"Expected a 'diseases' column. Found: {list(df.columns)[:20]} ...")
    disease_col = "diseases"
    symptom_cols = [c for c in df.columns if c != disease_col]

    symptom_counts: Dict[str, Dict[str, int]] = defaultdict(lambda: defaultdict(int))
    disease_row_counts: Dict[str, int] = defaultdict(int)
    symptom_to_diseases: Dict[str, set] = defaultdict(set)

    for row in df.itertuples(index=False):
        disease = str(getattr(row, disease_col)).strip().lower()
        if not disease:
            continue
        disease_row_counts[disease] += 1

        for col_name, val in zip(df.columns, row):
            if col_name == disease_col:
                continue
            try:
                if int(val) == 1:
                    sym = col_name.strip().lower()
                    symptom_counts[disease][sym] += 1
                    symptom_to_diseases[sym].add(disease)
            except Exception:
                continue

    return symptom_counts, disease_row_counts, symptom_to_diseases


def build_disease_docs(
    symptom_counts: Dict[str, Dict[str, int]],
    disease_row_counts: Dict[str, int],
) -> List[Dict[str, Any]]:
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
    return docs


def build_symptom_docs(symptom_to_diseases: Dict[str, set]) -> List[Dict[str, Any]]:
    return [
        {
            "_id": sym,
            "name": sym,
            "source": SOURCE_TAG,
            "diseases": sorted(diseases),
            "n_diseases": len(diseases),
        }
        for sym, diseases in symptom_to_diseases.items()
    ]


def main() -> None:
    load_dotenv(os.path.join(repo_root(), ".env"))
    mongo_uri = os.getenv("MONGODB_URI") or os.getenv("MONGO_URI")
    db_name = os.getenv("KAGGLE_DATASET_DB", "kaggle_dataset")

    if not mongo_uri:
        raise ValueError("Missing MONGODB_URI (or MONGO_URI) in .env")

    print("Resolving CSV...")
    csv_path = resolve_csv_path()
    print("CSV path:", csv_path)

    print("Reading CSV...")
    symptom_counts, disease_row_counts, symptom_to_diseases = parse_csv(csv_path)
    print(f"Unique diseases: {len(symptom_counts):,}, unique symptoms: {len(symptom_to_diseases):,}")

    disease_docs = build_disease_docs(symptom_counts, disease_row_counts)
    symptom_docs = build_symptom_docs(symptom_to_diseases)
    print(f"Prepared {len(disease_docs):,} disease docs, {len(symptom_docs):,} symptom docs.")

    db = KaggleDatasetDB(mongo_uri, db_name)
    try:
        db.replace_by_source("diseases", disease_docs, SOURCE_TAG)
        db.replace_by_source("symptoms", symptom_docs, SOURCE_TAG)
        db.ensure_indexes()
        print("Indexes created. DONE.")
    finally:
        db.close()


if __name__ == "__main__":
    main()
