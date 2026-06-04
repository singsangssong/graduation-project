# Agentic Middleware 최종 발표 스크립트

## 발표 목표

이 발표의 핵심 메시지는 "Agentic AI 환경에서는 트랜잭션 병목이 단순 성능 문제가 아니라 LLM 추론 비용 손실 문제로 확장된다"는 점이다. 프로젝트는 이를 Go 기반 미들웨어에서 QCFuse-style read fusion과 ATCC-style cost-aware commit arbitration으로 완화하는 구조를 제안하고 구현했다.

## 슬라이드별 진행 메모

1. Agentic Middleware
   - 다중 LLM 에이전트가 동시에 읽고 추론하고 commit을 시도할 때 공유 자원을 어떻게 보호할지 다루는 프로젝트라고 소개한다.
   - 발표의 키워드는 QCFuse, ATCC, Saga-style rollback이다.

2. 발표 흐름
   - 문제 정의, 관련 연구, 제안 구조, 구현, 실험, 데모 순서로 진행한다고 안내한다.

3. Agentic AI 시대의 워크로드 변화
   - 단일 LLM 호출이 아니라 여러 sub-agent가 병렬로 도구를 호출하는 상황이 늘어난다는 점을 강조한다.
   - 이 변화 때문에 DB 트랜잭션과 API 비용 문제가 함께 발생한다고 연결한다.

4. 문제 정의
   - 기존 OLTP는 트랜잭션이 짧지만, agentic workflow는 LLM reasoning 시간이 길다.
   - 긴 reasoning이 lock 대기와 rollback 비용으로 전이되는 것이 이 프로젝트의 핵심 문제다.

5. 관련 연구와 Gap
   - SagaLLM은 에이전트 workflow에 분산 트랜잭션 관점을 도입한다.
   - ATCC는 비용 기반 동시성 제어의 방향성을 준다.
   - QCFuse는 중복 접근을 줄이는 fusion 아이디어를 제공한다.
   - 이 프로젝트는 이를 졸업프로젝트 범위에서 실행 가능한 미들웨어 구조로 재구성했다.

6. 제안 아키텍처
   - Python agent layer와 Go middleware layer를 분리한 구조를 설명한다.
   - proto가 API의 단일 기준이고, Go gRPC 서버가 QCFuse/ATCC 로직을 담당한다고 말한다.

7. QCFuse-style 병렬 읽기
   - 동시에 들어온 read 요청을 짧은 window 동안 모아 한 번의 resource lookup으로 처리한다고 설명한다.
   - 100개 agent read가 1회 DB read로 줄어드는 흐름을 보여준다.

8. Lock-Free Reasoning
   - agent가 LLM reasoning을 수행하는 동안 DB lock을 잡지 않는 것이 핵심이다.
   - middleware는 read snapshot과 commit arbitration을 분리해서 lock 점유 시간을 줄인다.

9. ATCC-style Commit Arbitration
   - commit 요청을 token 사용량과 reasoning latency 기반 비용으로 정렬한다.
   - 가장 낮은 비용의 winner만 commit하고 나머지는 rollback signal을 받는다.
   - 무작위 rollback보다 매몰 비용 손실을 줄이는 점을 강조한다.

10. 시스템 구현
   - Go는 gRPC middleware와 metrics endpoint를 담당한다.
   - Python은 deterministic demo와 stress test client를 담당한다.
   - dashboard는 데모 영상에서 흐름을 보여주는 화면이다.

11. LLM 도구 호출 안정화
   - parameter 누락, 타입 오류, 통신 오류가 agentic system에서 실제로 자주 발생한다고 설명한다.
   - proto 기반 계약과 self-healing 재호출 흐름으로 안정성을 높였다고 말한다.

12. 실험 환경
   - 재고 1장, 다중 agent 동시 접속 상황으로 경합을 의도적으로 만든다.
   - 10/50/100/200 agent 수준에서 read fusion과 rollback 결과를 측정한다.

13. Deterministic Demo 결과
   - 고정된 5개 agent로 항상 같은 winner가 나오게 구성했다.
   - 발표 중 실시간 실행이 흔들려도 결과 설명이 가능한 안전한 데모다.

14. Stress Test 결과
   - 50개 agent에서는 DB read가 50회에서 1회로 줄어드는 98% 감소를 보여준다.
   - rollback 수가 늘어나는 것은 재고 1장 조건에서 정상적인 경합 결과라고 설명한다.

15. 경제적 효과
   - 무작위 rollback 방식은 이미 쓴 LLM 비용이 그대로 매몰된다.
   - ATCC-style arbitration은 비용이 낮은 transaction을 우선 commit해 손실을 줄인다.

16. 데모 시연 흐름
   - Go server 실행, Python demo 실행, dashboard 확인 순서로 보여준다.
   - 교수님에게는 "읽기 융합 -> reasoning 대기 -> commit 중재 -> rollback 신호" 순서를 화면에서 짚어준다.

17. 기여와 한계
   - 기여: Agentic AI workload에 맞춘 트랜잭션 미들웨어 프로토타입, QCFuse-style read fusion, ATCC-style arbitration, 데모 가능한 metrics.
   - 한계: 실제 LLM API 비용/장애 상황, 장기 실행 workflow, production-grade recovery는 추후 과제다.

18. 결론
   - Agentic AI 시대에는 middleware가 DB 안정성뿐 아니라 LLM 비용 보호까지 담당해야 한다는 결론으로 마무리한다.
   - 최종 문장은 "clear demo, measurable result, explainable architecture"로 가져가면 좋다.

## 최종평가 제출 체크리스트

- 발표자료: `agentic-middleware-final.key`
- 공유용 발표자료: `agentic-middleware-final.pdf`
- GitHub 링크: `https://github.com/singsangssong/graduation-project`
- 데모 화면: `dashboard.html`
- 데모 실행: Go server -> Python deterministic demo -> dashboard metrics 확인
- 최종보고서 초안: 발표 구조를 그대로 확장해서 작성
