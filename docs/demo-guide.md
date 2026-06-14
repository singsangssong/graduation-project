# 데모 가이드

## 데모 목표

데모는 다중 에이전트가 같은 자원에 접근할 때 다음 흐름이 연결되어 동작하는지
확인합니다.

1. 동시 읽기 요청이 미들웨어에 수집된다.
2. 에이전트는 읽기 이후 독립적으로 추론한다.
3. 커밋 후보가 하나의 중재 창에서 비교된다.
4. 승인된 Saga는 커밋되고 거절된 Saga는 보상된다.
5. 서버 재시작 이후 상태와 이벤트가 복구된다.

## 사전 준비

```sh
python3 -m pip install -r requirements.txt
```

Go와 Python 검증:

```sh
cd middleware-go
GOCACHE=/private/tmp/agenic-middleware-gocache go test ./...
```

```sh
python3 -m py_compile \
  agent-python/deterministic_demo.py \
  agent-python/saga_demo.py \
  agent-python/recovery_check.py
```

## 1. 결정적 읽기·중재 데모

터미널 1:

```sh
cd middleware-go
EXPERIMENT_MODE=full \
  GOCACHE=/private/tmp/agenic-middleware-gocache go run .
```

터미널 2:

```sh
cd agent-python
python3 deterministic_demo.py
```

이 데모는 고정된 에이전트 입력으로 읽기 배치와 커밋 후보 중재를 보여줍니다. 결과는
`agent-python/outputs/demo-results/` 또는 지정한 출력 디렉터리에 저장됩니다.

## 2. Saga 보상 데모

새로운 DB 경로를 지정해 서버를 시작합니다.

```sh
cd middleware-go
TICKET_STOCK=3 \
SAGA_DB_PATH=data/demo-saga.db \
ATCC_WINDOW_MS=500 \
GOCACHE=/private/tmp/agenic-middleware-gocache go run .
```

다른 터미널에서 실행합니다.

```sh
cd agent-python
python3 saga_demo.py
```

확인할 흐름:

```text
SAGA_STARTED
  -> STEP_COMPLETED
  -> VALIDATION_PASSED
  -> SAGA_COMMITTED

또는

SAGA_STARTED
  -> STEP_COMPLETED
  -> VALIDATION_PASSED
  -> SAGA_ABORTED
  -> COMPENSATION_COMPLETED
  -> SAGA_COMPENSATED
```

## 3. 재시작 복구 확인

Saga 데모가 끝난 뒤 Go 서버를 종료하고, 같은 `SAGA_DB_PATH`로 다시 실행합니다.

```sh
cd middleware-go
TICKET_STOCK=3 \
SAGA_DB_PATH=data/demo-saga.db \
GOCACHE=/private/tmp/agenic-middleware-gocache go run .
```

다른 터미널:

```sh
cd agent-python
python3 recovery_check.py
```

복구 확인은 기존 DB 상태가 유지되는 것이 목적입니다.

## 4. Live 비교 실행

```sh
./scripts/run_live_experiment.sh
```

실행기는 `baseline`, `qcfuse`, `full` 모드를 순서대로 실행하고 HTML 보고서를
생성합니다. 이는 기능 비교를 위한 짧은 실행이며 운영 성능 벤치마크가 아닙니다.

## 상태 관찰

서버 실행 중 다음 주소를 사용할 수 있습니다.

```text
http://localhost:8080/metrics
http://localhost:8080/resource?resource_id=flight_ticket_A
http://localhost:8080/events?saga_id=<saga_id>
```

## 반복 실행과 SQLite 상태

미들웨어는 복구를 위해 SQLite 상태를 보존합니다. 같은 DB 경로로 데모를 다시
실행하면 이전 커밋으로 감소한 재고와 Saga 기록이 남아 있습니다.

새로운 독립 데모를 실행하려면 새 DB 경로를 사용하는 방식을 권장합니다.

```sh
SAGA_DB_PATH=data/demo-saga-2.db
```

기존 DB를 제거할 때는 해당 DB를 사용하는 미들웨어 프로세스가 종료되어 있는지 먼저
확인해야 합니다.
