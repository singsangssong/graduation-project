# 10-Minute Presentation And Live Experiment Scenario

## Before The Presentation

Run the full repeated profile in advance:

```sh
./scripts/run_paper_experiment.sh
```

Keep its `summary.csv`, `reliability.json`, and `report.html` available as the
primary evaluation evidence. Immediately before presenting, run:

```sh
./scripts/run_live_experiment.sh
```

Confirm that the generated report opens and the reliability result is `8/8`.
The live command is safe to run again because it uses isolated ports and SQLite
files.

## Ten-Minute Timeline

### 0:00-0:50 — Research Background

**Slide:** Agentic AI and AIOS perspective

**Narration:**

> 여러 LLM Agent가 도구와 공유 자원에 동시에 접근하는 환경에서는 단순한
> API 호출을 넘어 운영체제와 같은 스케줄링 및 자원 보호 계층이 필요합니다.
> 특히 Agent는 추론에 수 초에서 수 분을 사용하고 토큰 비용도 소비하기
> 때문에, 기존의 짧은 OLTP 트랜잭션과 다른 동시성 문제가 발생합니다.

### 0:50-1:40 — Problem Definition

**Slide:** Existing OLTP versus agentic transaction

**Narration:**

> 여러 Agent가 동일한 티켓을 조회하고 각자 긴 추론을 마친 뒤 커밋하면,
> 중복 조회 I/O와 커밋 경합이 발생합니다. 기존의 무작위 롤백은 이미 많은
> 추론 비용을 사용한 Agent까지 제거해 매몰 비용을 증가시킵니다.

### 1:40-2:50 — Related Work And Gap

**Slide:** SagaLLM, ATCC, QCFuse, AIOS comparison

**Narration:**

> SagaLLM은 Agent workflow의 보상 트랜잭션을, ATCC는 비용 인지형 동시성
> 제어를, QCFuse는 중복 계산 융합을 제안합니다. 본 프로젝트는 논문 전체를
> 재현했다고 주장하지 않고, 각 핵심 아이디어를 공유 자원 미들웨어에
> 결합한 SagaLLM-compatible, ATCC-style, QCFuse-style 시스템입니다.

### 2:50-4:10 — Proposed Architecture

**Slide:** Python Agent → gRPC → Go Middleware → SQLite/resource

**Narration:**

> Python Agent 계층과 Go 미들웨어를 gRPC로 분리했습니다. 읽기 요청은
> resource와 intent 기준으로 융합하고, 추론 중에는 DB lock을 점유하지
> 않습니다. 커밋 시점에는 토큰과 추론 시간을 비용 신호로 사용해 가장
> 보호 가치가 높은 요청을 선택합니다. 패배한 Saga는 역순 보상되고 모든
> 상태와 이벤트는 SQLite에 저장됩니다.

### 4:10-5:10 — Implementation Details

**Slide:** QCFuse, ATCC score, persistent Saga lifecycle

**Narration:**

> QCFuse window는 동일 요청 N개를 한 번의 logical DB read로 처리합니다.
> ATCC 점수는 현재 졸업프로젝트 범위에서 토큰 가중치와 추론 지연 가중치의
> 합으로 단순화했습니다. Saga는 ACTIVE, VALIDATED, COMMITTED 또는
> COMPENSATED 상태로 전이되고, 지원되지 않는 보상은 명시적으로
> COMPENSATION_FAILED가 됩니다.

### 5:10-6:50 — Live Experiment

**Action:** Run:

```sh
./scripts/run_live_experiment.sh
```

**Narration while running:**

> 지금 동일한 seeded workload를 baseline, QCFuse-only, full 모드에
> 순서대로 실행하고 있습니다. Baseline은 개별 logical read와 도착순
> 커밋을 사용합니다. QCFuse-only는 읽기만 융합하고, full 모드는 읽기
> 융합과 비용 인지 커밋, Saga 신뢰성 검사를 모두 적용합니다.

**Action:** Open the generated `report.html`.

> 50명 환경에서 baseline의 I/O 감소율은 0%이고 QCFuse와 full은 98%입니다.
> 도착순 모드는 최고 비용 Agent 보호율이 낮지만 full 모드는 100%입니다.
> 마지막으로 보상, 중복 보상 방지, 실패 탐지, 재시작 복구를 포함한
> 신뢰성 검사도 8개 모두 통과했습니다.

### 6:50-8:20 — Repeated Experiment Results

**Slide:** Paper-profile mean and standard deviation charts

**Narration:**

> 방금 결과는 시연 안정성을 위한 짧은 live profile입니다. 최종 평가는
> 10, 50, 100, 200명의 Agent를 각 비교군에서 5회 반복한 paper profile의
> 평균과 표준편차를 사용합니다. 여기서 중요한 결과는 logical I/O 감소,
> p95 지연, 처리량, 최고 비용 요청 보호율을 동시에 비교했다는 점입니다.

### 8:20-9:15 — Reliability And Recovery

**Slide:** Saga event timeline and restart recovery

**Narration:**

> 성능뿐 아니라 correctness도 분리해 평가했습니다. 패배한 Saga의 예약은
> 정확히 한 번 복구되며, 같은 보상을 다시 요청해도 재고가 두 번 증가하지
> 않습니다. 미들웨어를 종료하고 같은 SQLite DB로 재시작한 뒤에도 Saga
> 상태, 이벤트 타임라인, 자원 재고가 모두 복구됩니다.

### 9:15-10:00 — Conclusion And Limitations

**Slide:** Contributions and future work

**Narration:**

> 본 프로젝트의 기여는 Agent workflow 신뢰성과 공유 자원 동시성 제어를
> 하나의 미들웨어에서 결합하고, 반복 가능한 실험 환경까지 제공한 것입니다.
> 현재 실험은 ticket-stock simulation과 logical I/O counter를 사용하며,
> ATCC도 RL 전체 구현이 아닌 비용 기반 우선순위 구현입니다. 향후 실제
> PostgreSQL lock wait와 실제 LLM token metering으로 확장하겠습니다.

## Failure-Safe Demo Plan

If the live command cannot run, immediately open the latest pre-generated
`report.html` and say:

> 라이브 환경의 네트워크 또는 포트 문제에 대비해 동일 명령으로 사전
> 생성한 결과를 보여드리겠습니다. 원시 CSV와 반복 실험 JSON도 함께
> 보존되어 있어 결과를 재현하고 검토할 수 있습니다.

Do not spend presentation time debugging a local port or dependency issue.
