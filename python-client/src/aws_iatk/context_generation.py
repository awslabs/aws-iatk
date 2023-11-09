import uuid
from datetime import datetime

def eventbridge_event_context(event_dict: dict) -> dict:
    """
    Function that adds default values for event bridge specific event key/value pairs if they don't exist already

    Parameters
    ----------
    event_dict : dict
    dictionary that represents the event that will be passed into event bridge
    
    Returns
    -------
    dict
        same dictionary with the event bridge specfic values added 
    """
    if not event_dict.get("version"):
        event_dict["version"] = "0"
    if not event_dict.get("id") or event_dict.get("id") == "":
        event_dict["id"] = str(uuid.uuid4())
    if not event_dict.get("detail-type") or event_dict.get("detail-type") == "":
        event_dict["detail-type"] = "detail-type"
    if not event_dict.get("source") or event_dict.get("source") == "":
        event_dict["source"] = "source"
    if not event_dict.get("account"):
        event_dict["account"] = "123456789101"
    if not event_dict.get("time"):
        event_dict["time"] = datetime.now().strftime("%Y-%m-%dT%H:%M:%S")
    if not event_dict.get("region"):
        event_dict["region"] = "us-east-1"
    if not event_dict.get("resources"):
        event_dict["resources"] = []

    return event_dict