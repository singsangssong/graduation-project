import grpc
import middleware_pb2
import middleware_pb2_grpc
import concurrent.futures
import random
import time

def simulate_agent_full_workflow(agent_id: str):
    # ATCC 논문 반영: 에이전트 트랜잭션의 예측 불가능한 추론 비용(토큰)과 지연 시간 세팅
    token_usage = random.randint(500, 5000)
    latency_sec = round(random.uniform(2.0, 10.0), 2)
    resource_id = "flight_ticket_A"
    
    start_time = time.time()
    
    try:
        # 미들웨어 고속도로(gRPC) 탑승
        channel = grpc.insecure_channel('localhost:50051')
        stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
        
        # =======================================================
        # [Phase 1] QCFuse (Query-Centric Cache Fusion) 조회 요청
        # 200명의 에이전트가 동일한 자원을 동시에 조회 (DB 병목 방어 테스트)
        # =======================================================
        read_req = middleware_pb2.ReadRequest(agent_id=agent_id, resource_id=resource_id)
        read_resp = stub.ReadResource(read_req) # 향후 미들웨어에서 버퍼링 적용 예정
        
        # =======================================================
        # [Phase 2] Agentic Reasoning (Lock-Free 대기)
        # ATCC 논문에서 강조한 '긴 추론 시간' 모사. 
        # DB Lock을 쥐지 않은 상태로 파이썬 단에서 독자적으로 연산 대기
        # =======================================================
        time.sleep(latency_sec) 
        
        # =======================================================
        # [Phase 3] ATCC (Adaptive Concurrency Control) 커밋 요청
        # 추론을 마친 에이전트들이 매몰 비용 데이터를 들고 커밋 경합 시도
        # =======================================================
        commit_req = middleware_pb2.CommitRequest(
            agent_id=agent_id,
            resource_id=resource_id,
            action_value=1,
            accumulated_tokens=token_usage,
            inference_latency_sec=latency_sec
        )
        commit_resp = stub.CommitTransaction(commit_req)
        
        total_time = round(time.time() - start_time, 2)
        
        # 결과 반환
        if commit_resp.is_rolled_back:
            return {"status": "rollback", "agent": agent_id, "tokens": token_usage, "latency": latency_sec}
        else:
            return {"status": "success", "agent": agent_id, "tokens": token_usage, "latency": latency_sec}

    except grpc.RpcError as e:
        return {"status": "error", "agent": agent_id, "error_msg": e.details()}

def run_advanced_stress_test(thread_count: int):
    print(f"\n{'='*50}")
    print(f"🚀 [논문 검증 테스트 시작] 동시 접속 에이전트: {thread_count}명")
    print(f"{'='*50}")
    
    results = {"success": 0, "rollback": 0, "error": 0}
    
    # concurrent.futures로 논문 환경과 동일한 고밀도 동시성 부하 발생
    with concurrent.futures.ThreadPoolExecutor(max_workers=thread_count) as executor:
        futures = [executor.submit(simulate_agent_full_workflow, f"Agent_{i:03d}") for i in range(1, thread_count + 1)]
        
        for future in concurrent.futures.as_completed(futures):
            res = future.result()
            results[res["status"]] += 1
            
    print(f"📊 [테스트 결과 요약] 총 {thread_count}건의 Agentic Transaction 처리")
    print(f" - 커밋 승인(Success)   : {results['success']}건")
    print(f" - ATCC 롤백(Rollback)  : {results['rollback']}건")
    print(f" - 통신 에러(Timeout)   : {results['error']}건")

if __name__ == "__main__":
    # 스레드 수를 10, 50, 100, 200으로 늘려가며 Scalability 테스트
    test_cases = [10, 50, 100, 200]
    
    for count in test_cases:
        run_advanced_stress_test(count)
        # 다음 테스트 전 커넥션 풀 안정을 위한 대기
        time.sleep(3)