import grpc
import middleware_pb2
import middleware_pb2_grpc
import concurrent.futures
import random
import time

def simulate_agent_request(agent_id: str):
    # 1. 교수님 피드백 반영: 다양한 토큰 값 및 대기 시간 세분화 (난수 생성)
    token_usage = random.randint(500, 5000)
    latency_sec = round(random.uniform(2.0, 10.0), 2)
    
    try:
        # 미들웨어 서버 연결
        channel = grpc.insecure_channel('localhost:50051')
        stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
        
        request = middleware_pb2.CommitRequest(
            agent_id=agent_id,
            resource_id="flight_ticket_A",
            action_value=1,
            accumulated_tokens=token_usage,
            inference_latency_sec=latency_sec
        )
        
        # 2. 커밋 요청 발송
        response = stub.CommitTransaction(request)
        
        # 결과 반환 (이슈 트래킹을 위해 상태 분류)
        if response.is_rolled_back:
            return {"status": "rollback", "agent": agent_id, "tokens": token_usage, "saved_cost": response.saved_cost_usd}
        else:
            return {"status": "success", "agent": agent_id, "tokens": token_usage}
            
    except grpc.RpcError as e:
        # gRPC 통신 에러 발생 시 트래킹을 위해 에러 반환
        return {"status": "error", "agent": agent_id, "error_msg": e.details()}

def run_stress_test(thread_count: int):
    print(f"\n{'='*40}")
    print(f"🚀 [테스트 시작] 동시 접속 에이전트: {thread_count}명")
    print(f"{'='*40}")
    
    results = {"success": 0, "rollback": 0, "error": 0}
    
    # 3. 교수님 피드백 반영: 스레드 수 제어 및 동시 실행 (concurrent.futures 활용)
    with concurrent.futures.ThreadPoolExecutor(max_workers=thread_count) as executor:
        # 지정된 스레드 수만큼 에이전트 요청 동시 생성
        futures = [executor.submit(simulate_agent_request, f"Agent_{i:03d}") for i in range(1, thread_count + 1)]
        
        for future in concurrent.futures.as_completed(futures):
            res = future.result()
            results[res["status"]] += 1
            
    # 테스트 결과 요약 출력
    print(f"📊 테스트 결과 요약 (총 {thread_count}건)")
    print(f" - 커밋 성공   : {results['success']}건")
    print(f" - 롤백(보상)  : {results['rollback']}건")
    print(f" - 통신 에러   : {results['error']}건")
    if results['error'] > 0:
        print(" ⚠️ [이슈 트래킹 포인트] 통신 에러가 발생했습니다. gRPC 커넥션 한계 또는 타임아웃을 점검해야 합니다.")

if __name__ == "__main__":
    # 4. 스레드 수를 10, 50, 100, 200으로 점진적으로 늘려가며 부하 테스트 진행
    test_cases = [10, 50, 100, 200]
    
    for count in test_cases:
        run_stress_test(count)
        # 다음 테스트 전 서버가 안정을 찾도록 2초 대기
        time.sleep(2)