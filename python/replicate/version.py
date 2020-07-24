from typing import Dict


def attach_version(data: Dict) -> Dict:
    data = data.copy()
    data["metadata_version"] = "1"
    return data
