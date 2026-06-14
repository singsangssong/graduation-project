# Graduation Project Final Presentation Strategy

## 발표 제목 후보

**Agentic Middleware: 다중 LLM 에이전트 트랜잭션을 위한 QCFuse + ATCC 기반 미들웨어**

## 현재 목차 보강안

### 1. 서론

현재 목차는 좋지만, 아래 3가지를 명시하면 발표 설득력이 커진다.

- **Agentic workload의 변화**
  - 단일 LLM 호출에서 벗어나, 하나의 목표를 여러 sub-agent가 동시에 탐색/추론/도구 호출하는 구조로 이동한다.
  - Concurrent Modular Agent 논문은 여러 LLM 기반 모듈이 완전히 비동기적으로 동작하는 구조를 제안하며, autonomous agent 구조에서 concurrent processing이 중요해진다는 점을 보여준다.
- **기존 DB 트랜잭션 가정의 붕괴**
  - 전통적인 OLTP는 짧고 예측 가능한 transaction을 전제로 한다.
  - agentic transaction은 LLM 추론으로 인해 실행 시간이 길고, SQL/API 접근 순서가 런타임에 바뀌며, 충돌 시 abort 비용이 매우 크다.
- **본 프로젝트의 문제 정의**
  - 긴 LLM reasoning 시간이 DB lock 점유/대기 시간으로 전이된다.
  - rollback이 발생하면 이미 소비한 token, API 호출, reasoning latency가 매몰 비용이 된다.
  - 다중 agent가 동일 자원을 조회할 때 DB read amplification이 발생한다.

### 2. 관련 연구 및 논문 정리

#### 2.1 Concurrent Modular Agent

- 여러 LLM 기반 모듈이 비동기적으로 병렬 작동하는 agent framework.
- 본 프로젝트에서는 "왜 다중 agent 동시 실행이 일반화되는가"의 배경 근거로 사용한다.
- 한계: concurrency 자체를 다루지만, 공유 DB 자원에 대한 transaction conflict나 rollback cost 문제를 직접 해결하지 않는다.

#### 2.2 AIOS

- LLM agent를 위한 operating system 관점의 구조를 제안한다.
- scheduler, memory manager, tool manager 등의 구성요소를 통해 agent 실행을 시스템 레벨에서 관리하는 방향을 보여준다.
- 본 프로젝트에서는 "agent를 application 내부 로직이 아니라 middleware/OS-like layer에서 제어할 수 있다"는 설계 배경으로 사용한다.

#### 2.3 SagaLLM

- multi-agent LLM planning에 Saga pattern을 적용하여 checkpoint, validation, compensation, rollback 구조를 제공한다.
- 본 프로젝트에서는 "agent workflow에도 transaction guarantee가 필요하다"는 근거로 사용한다.
- 한계:
  - workflow consistency와 compensation에 초점이 있다.
  - 다중 agent가 동일 DB resource를 동시에 접근할 때 생기는 lock contention, abort priority, sunk cost-aware rollback 문제는 충분히 다루지 않는다.

#### 2.4 ATCC

- agentic transaction은 long lifetime, irregular interval, non-deterministic access pattern을 가지며 전통적 concurrency control 가정을 깨뜨린다고 분석한다.
- abort가 발생하면 단순 DB 작업뿐 아니라 multi-round LLM reasoning과 token cost가 함께 손실된다.
- ATCC는 execution time, reasoning cost 등을 고려해 transaction priority를 산정하고, 충돌 상황에서 비용이 큰 transaction을 우선 보호하는 방향을 제안한다.
- 본 프로젝트에서는 이를 단순화하여:
  - accumulated tokens
  - inference latency
  - cost formula
  - leaderboard sorting
  - winner commit / loser rollback signal
  구조로 구현한다.

#### 2.5 QCFuse

- QCFuse는 RAG inference에서 query-centric KV cache fusion을 통해 중복 계산을 줄이고 latency를 낮추는 시스템이다.
- 원 논문은 LLM KV cache와 token recomputation을 다루지만, 본 프로젝트에서는 핵심 아이디어를 "동일 자원에 대한 query fusion"으로 재해석한다.
- 구현상 차이:
  - 원 논문: RAG/LLM KV cache fusion.
  - 본 프로젝트: Go channel window 기반 read request fusion.
- 발표에서는 "QCFuse-inspired query fusion" 또는 "QCFuse-style read fusion"이라고 표현하는 것이 정확하다.

### 3. 제안 아키텍처

현재 목차에 아래를 추가하면 좋다.

- **Architecture Boundary**
  - Python layer: agent simulation, AutoGen/LangGraph-style tool invocation, stress test.
  - Go middleware layer: gRPC API, QCFuse scheduler, ATCC scheduler, metrics endpoint.
  - DB/resource layer: 현재는 ticket stock simulation, 향후 MySQL/Redis 연동 가능.
- **Why Go?**
  - channel 기반 queue와 goroutine scheduler로 concurrent middleware prototype을 간결하게 구현할 수 있다.
- **Why gRPC?**
  - agent layer와 middleware layer 사이의 typed contract를 제공한다.
  - protobuf를 통해 Python/Go 간 API mismatch를 줄인다.

### 4. 시스템 구현

기존 목차에 아래 구현 항목을 추가한다.

- **Canonical Proto Contract**
  - `proto/middleware.proto`를 단일 API 계약으로 사용한다.
  - Python/Go generated code를 같은 proto에서 재생성한다.
- **Metrics API**
  - `/metrics`: read requests, saved DB reads, commit requests, rollback count, winner, saved cost.
  - `/reset`: 데모 실행 전 metric reset.
- **Presentation Dashboard**
  - 데모 영상에서 QCFuse -> ATCC 흐름을 시각적으로 보여준다.
- **Deterministic Demo**
  - 랜덤 실험과 별개로 항상 같은 winner/rollback 결과가 나오는 발표용 스크립트를 제공한다.

### 5. 실험 및 성능 평가

주의: 현재 발표 수치로 적어둔 `I/O 부하 99% 감소`, `850TPS`, `매몰 비용 73% 방어`는 반드시 실제 측정 결과와 연결해야 한다.

권장 실험 구성:

1. **Deterministic demo**
   - 5 agents
   - expected winner: Agent-A
   - saved DB reads: 4
   - rollback: 4
   - saved cost: 17.70 USD equivalent

2. **Scalability stress test**
   - 10, 50, 100, 200 agents
   - metrics:
     - read requests
     - fused read batches
     - saved DB reads
     - commit requests
     - approved commits
     - rolled back commits
     - total saved cost
     - end-to-end elapsed time

3. **Baseline comparison**
   - without QCFuse: DB read count = N
   - with QCFuse: DB read count ≈ number of fusion batches
   - without ATCC: random winner may waste high sunk-cost agent
   - with ATCC: highest sunk-cost agent wins

### 6. 데모 시연

권장 데모 흐름:

1. `python3 agent-python/experiment_runner.py --profile live` 실행
2. baseline/QCFuse/full 비교 진행 상황 확인
3. 생성된 `report.html` 열기
4. QCFuse logical I/O 감소율과 full 모드 winner protection 100% 설명
5. Saga reliability 8/8 및 restart recovery 설명
6. 시간이 부족하거나 실행 실패 시 사전 생성한 paper/live 결과를 사용

구체적인 10분 발표 흐름과 대본은 `docs/presentation_10min_scenario.md`를
기준으로 한다.

### 7. 결론

결론에는 아래 3가지를 넣는다.

- Agentic AI 환경에서는 DB transaction이 단순 데이터 정합성 문제가 아니라 token/API/reasoning cost 보존 문제로 확장된다.
- 본 프로젝트는 QCFuse-style read fusion과 ATCC-style cost-aware arbitration을 middleware layer에서 결합했다.
- 향후에는 실제 DB/Redis 연동, LangGraph/AutoGen end-to-end workflow, real LLM token metering, baseline stress benchmark를 확장한다.

## 최종 발표에서 추가하면 좋은 슬라이드

1. **Agentic Transaction이 기존 OLTP와 다른 이유**
   - 실행 시간, 접근 패턴, abort cost 비교 표.
2. **Related Work Gap**
   - SagaLLM / ATCC / QCFuse / 본 프로젝트 비교 표.
3. **Architecture Diagram**
   - Python Agent -> gRPC -> Go Middleware -> QCFuse/ATCC -> DB.
4. **Algorithm Slide**
   - QCFuse window batching pseudocode.
   - ATCC sunk cost scoring pseudocode.
5. **Demo Result Slide**
   - deterministic demo 결과 화면.
6. **Stress Test Slide**
   - 10/50/100/200 agent 결과 그래프.
7. **Limitations**
   - 현재 DB는 simulation.
   - cost formula는 simplified proxy.
   - AutoGen/LangGraph full workflow는 optional scenario.

## 교수님 질문 대비

### Q1. QCFuse 논문은 KV cache fusion인데, 왜 DB read fusion에 사용했나요?

원 논문의 핵심은 query-centric 관점에서 중복 context computation을 줄이는 것이다. 본 프로젝트는 이를 LLM KV cache가 아닌 middleware read path에 적용해, 동일 resource query를 짧은 window에서 fusion하는 방식으로 재해석했다. 따라서 정확한 표현은 QCFuse 자체 구현이 아니라 QCFuse-inspired read fusion이다.

### Q2. ATCC와 다른 점은 무엇인가요?

ATCC 원 논문은 RL 기반 phase-aware concurrency control과 priority locking을 제안한다. 본 프로젝트는 졸업프로젝트 범위에 맞춰 priority scheduling의 핵심 아이디어를 단순화했고, token/latency 기반 sunk cost score로 commit winner를 결정한다.

### Q3. 왜 middleware layer인가요?

Agent framework 내부에 transaction control을 넣으면 agent 구현마다 반복된다. middleware layer로 분리하면 Python/AutoGen/LangGraph와 같은 agent runtime과 DB/resource layer 사이에 공통 transaction guard를 제공할 수 있다.

### Q4. 현재 실험 수치의 한계는?

현재 수치는 controlled simulation에서 측정된다. 따라서 최종 보고서에는 "simulated ticket stock workload"라고 명확히 적고, 실제 DB 연동은 향후 과제로 제시하는 것이 안전하다.

## 참고 논문 URL

- QCFuse: https://arxiv.org/pdf/2604.08585
- Concurrent Modular Agent: https://arxiv.org/pdf/2508.19042
- SagaLLM: https://arxiv.org/pdf/2503.11951
- AIOS: https://arxiv.org/pdf/2403.16971
- ATCC: https://arxiv.org/pdf/2603.13906
