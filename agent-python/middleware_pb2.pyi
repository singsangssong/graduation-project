from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ReadRequest(_message.Message):
    __slots__ = ("agent_id", "resource_id", "intent", "saga_id")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    INTENT_FIELD_NUMBER: _ClassVar[int]
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    resource_id: str
    intent: str
    saga_id: str
    def __init__(self, agent_id: _Optional[str] = ..., resource_id: _Optional[str] = ..., intent: _Optional[str] = ..., saga_id: _Optional[str] = ...) -> None: ...

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
    __slots__ = ("agent_id", "resource_id", "action_value", "accumulated_tokens", "inference_latency_sec", "saga_id")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_VALUE_FIELD_NUMBER: _ClassVar[int]
    ACCUMULATED_TOKENS_FIELD_NUMBER: _ClassVar[int]
    INFERENCE_LATENCY_SEC_FIELD_NUMBER: _ClassVar[int]
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    resource_id: str
    action_value: int
    accumulated_tokens: int
    inference_latency_sec: float
    saga_id: str
    def __init__(self, agent_id: _Optional[str] = ..., resource_id: _Optional[str] = ..., action_value: _Optional[int] = ..., accumulated_tokens: _Optional[int] = ..., inference_latency_sec: _Optional[float] = ..., saga_id: _Optional[str] = ...) -> None: ...

class CommitResponse(_message.Message):
    __slots__ = ("success", "is_rolled_back", "message", "saved_cost_usd", "saga_status")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    IS_ROLLED_BACK_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    SAVED_COST_USD_FIELD_NUMBER: _ClassVar[int]
    SAGA_STATUS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    is_rolled_back: bool
    message: str
    saved_cost_usd: float
    saga_status: str
    def __init__(self, success: bool = ..., is_rolled_back: bool = ..., message: _Optional[str] = ..., saved_cost_usd: _Optional[float] = ..., saga_status: _Optional[str] = ...) -> None: ...

class BeginSagaRequest(_message.Message):
    __slots__ = ("agent_id", "goal", "context")
    class ContextEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    GOAL_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    goal: str
    context: _containers.ScalarMap[str, str]
    def __init__(self, agent_id: _Optional[str] = ..., goal: _Optional[str] = ..., context: _Optional[_Mapping[str, str]] = ...) -> None: ...

class RegisterSagaStepRequest(_message.Message):
    __slots__ = ("saga_id", "step_id", "action", "result", "compensation_action")
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    STEP_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    COMPENSATION_ACTION_FIELD_NUMBER: _ClassVar[int]
    saga_id: str
    step_id: str
    action: str
    result: str
    compensation_action: str
    def __init__(self, saga_id: _Optional[str] = ..., step_id: _Optional[str] = ..., action: _Optional[str] = ..., result: _Optional[str] = ..., compensation_action: _Optional[str] = ...) -> None: ...

class ValidateSagaRequest(_message.Message):
    __slots__ = ("saga_id",)
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    saga_id: str
    def __init__(self, saga_id: _Optional[str] = ...) -> None: ...

class AbortSagaRequest(_message.Message):
    __slots__ = ("saga_id", "reason")
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    saga_id: str
    reason: str
    def __init__(self, saga_id: _Optional[str] = ..., reason: _Optional[str] = ...) -> None: ...

class GetSagaStateRequest(_message.Message):
    __slots__ = ("saga_id",)
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    saga_id: str
    def __init__(self, saga_id: _Optional[str] = ...) -> None: ...

class SagaStep(_message.Message):
    __slots__ = ("step_id", "action", "result", "compensation_action", "status")
    STEP_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    COMPENSATION_ACTION_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    step_id: str
    action: str
    result: str
    compensation_action: str
    status: str
    def __init__(self, step_id: _Optional[str] = ..., action: _Optional[str] = ..., result: _Optional[str] = ..., compensation_action: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class SagaState(_message.Message):
    __slots__ = ("saga_id", "agent_id", "goal", "status", "steps", "validation_message", "abort_reason", "created_at_unix_seconds", "updated_at_unix_seconds")
    SAGA_ID_FIELD_NUMBER: _ClassVar[int]
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    GOAL_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    STEPS_FIELD_NUMBER: _ClassVar[int]
    VALIDATION_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ABORT_REASON_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_UNIX_SECONDS_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_UNIX_SECONDS_FIELD_NUMBER: _ClassVar[int]
    saga_id: str
    agent_id: str
    goal: str
    status: str
    steps: _containers.RepeatedCompositeFieldContainer[SagaStep]
    validation_message: str
    abort_reason: str
    created_at_unix_seconds: int
    updated_at_unix_seconds: int
    def __init__(self, saga_id: _Optional[str] = ..., agent_id: _Optional[str] = ..., goal: _Optional[str] = ..., status: _Optional[str] = ..., steps: _Optional[_Iterable[_Union[SagaStep, _Mapping]]] = ..., validation_message: _Optional[str] = ..., abort_reason: _Optional[str] = ..., created_at_unix_seconds: _Optional[int] = ..., updated_at_unix_seconds: _Optional[int] = ...) -> None: ...

class SagaResponse(_message.Message):
    __slots__ = ("success", "message", "saga")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    SAGA_FIELD_NUMBER: _ClassVar[int]
    success: bool
    message: str
    saga: SagaState
    def __init__(self, success: bool = ..., message: _Optional[str] = ..., saga: _Optional[_Union[SagaState, _Mapping]] = ...) -> None: ...
