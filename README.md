# Context-Aware Symptom Info Search  
**CS 125 Project**

A context-aware health information search + recommendation prototype that helps users find **trustworthy next-step guidance** for symptom searches without information overload.

> This system is **not** a medical diagnosis tool and does **not** replace professional care.

---

## Team Information
**Team Name:** The Leet Koders  
**Team Members (Name, UCINetID):**
- Thomas Nguyen — `thomaln5`
- Lauren Gabrielle Yraola — `lyraola`
- Kyle Yin Xu — `kyinxu`

---

## Project Overview
People often search symptoms online when they feel unwell, but results can be overwhelming, misleading, or not prioritized for their situation. This project builds a **context-aware symptom search system** that ranks helpful information based on implicit context signals like:

- time of day  
- approximate location  
- session/search history  
- user interaction preferences  

The system prioritizes next-step guidance (monitoring, escalation, nearby care, etc.) rather than just returning generic articles.

---

## Key Features (Planned)
- Symptom-based search over a curated knowledge base
- Ranked information cards such as:
  - Symptom overview
  - Monitoring checklist (what to track)
  - “When to seek care” guidelines
  - Red flags / warning signs
  - Self-care guidance (non-diagnostic)
  - Questions to ask a clinician
  - Suggested next searches / follow-up prompts
  - Nearby urgent care recommendations *(optional)*
- Context-aware reranking using:
  - time of day
  - approximate location (IP-based)
  - session history
- “Why this was recommended” explanation for each result

---

To avoid medical and safety risks, this system will **not**:
- diagnose conditions
- replace a healthcare professional
- connect to medical records or wearables
- track precise GPS location
- predict emergencies or provide live wait times

---

## High-Level System Concept
1. User selects or searches a symptom  
2. System retrieves candidate resources from a trusted knowledge base  
3. Context signals are collected (time/location/session history)  
4. A lightweight personal model boosts result types the user prefers  
5. Results are ranked and displayed with short explanations

---

## Data Sources
This project uses:
- **User symptom input** (plus optional details like duration/progression)
- **Curated health info knowledge base** (from reputable public sources)
- **Context signals**
  - time of day
  - approximate location (IP-based)
  - session history
- Nearby care dataset/API (distance, rating, open-now status)

