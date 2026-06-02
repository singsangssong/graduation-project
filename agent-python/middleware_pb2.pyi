from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ReadRequest(_message.Message):
    __slots__ = ("agent_id", "resource_id", "intent")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    INTENT_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    resource_id: str
    intent: str
    def __init__(self, agent_id: _Optional[str] = ..., resource_id: _Optional[str] = ..., intent: _Optional[str] = ...) -> None: ...

class ReadResponse(_message.Message):
    __slots__ = ("success", "data", "message")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    success: bool
    data: str
    message: str
    def __init__(self, success: bool = ..., data: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class CommitRequest(_message.Message):
    __slots__ = ("agent_id", "resource_id", "action_value", "accumulated_tokens", "inference_latency_sec")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_VALUE_FIELD_NUMBER: _ClassVar[int]
    ACCUMULATED_TOKENS_FIELD_NUMBER: _ClassVar[int]
    INFERENCE_LATENCY_SEC_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    resource_id: str
    action_value: int
    accumulated_tokens: int
    inference_latency_sec: float
    def __init__(self, agent_id: _Optional[str] = ..., resource_id: _Optional[str] = ..., action_value: _Optional[int] = ..., accumulated_tokens: _Optional[int] = ..., inference_latency_sec: _Optional[float] = ...) -> None: ...

class CommitResponse(_message.Message):
    __slots__ = ("success", "is_rolled_back", "message", "saved_cost_usd")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    IS_ROLLED_BACK_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    SAVED_COST_USD_FIELD_NUMBER: _ClassVar[int]
    success: bool
    is_rolled_back: bool
    message: str
    saved_cost_usd: float
    def __init__(self, success: bool = ..., is_rolled_back: bool = ..., message: _Optional[str] = ..., saved_cost_usd: _Optional[float] = ...) -> None: ...
