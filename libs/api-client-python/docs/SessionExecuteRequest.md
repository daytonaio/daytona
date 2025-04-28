# SessionExecuteRequest

## Properties

| Name          | Type     | Description                                                                     | Notes      |
| ------------- | -------- | ------------------------------------------------------------------------------- | ---------- |
| **command**   | **str**  | The command to execute                                                          |
| **run_async** | **bool** | Whether to execute the command asynchronously                                   | [optional] |
| **var_async** | **bool** | Deprecated: Use runAsync instead. Whether to execute the command asynchronously | [optional] |

## Example

```python
from daytona_api_client.models.session_execute_request import SessionExecuteRequest

# TODO update the JSON string below
json = "{}"
# create an instance of SessionExecuteRequest from a JSON string
session_execute_request_instance = SessionExecuteRequest.from_json(json)
# print the JSON string representation of the object
print(SessionExecuteRequest.to_json())

# convert the object into a dict
session_execute_request_dict = session_execute_request_instance.to_dict()
# create an instance of SessionExecuteRequest from a dict
session_execute_request_from_dict = SessionExecuteRequest.from_dict(session_execute_request_dict)
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
