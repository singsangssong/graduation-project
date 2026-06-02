import grpc
import middleware_pb2
import middleware_pb2_grpc

# 1. 미들웨어 서버(Go)와 연결하는 통로(Channel) 생성
def get_middleware_stub():
    # Go 서버가 열어둔 50051 포트로 연결해!
    channel = grpc.insecure_channel('localhost:50051')
    stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
    return stub

# 2. Agentic AI(AutoGen)가 실제로 사용할 "결제 도구(Tool)" 함수
def request_commit_tool(agent_id: str, resource_id: str, accumulated_tokens: int, inference_latency_sec: float) -> str:
    """
    에이전트가 결제를 요청할 때 호출하는 함수입니다.
    이 함수가 실행되면 Go 미들웨어로 gRPC 요청이 날아갑니다.
    """
    print(f"[{agent_id}] 미들웨어에 커밋을 요청합니다... (누적 토큰: {accumulated_tokens}, 대기시간: {inference_latency_sec}초)")
    
    stub = get_middleware_stub()
    
    # Go 서버(middleware.proto)에서 정의한 형식대로 데이터 포장하기
    request = middleware_pb2.CommitRequest(
        agent_id=agent_id,
        resource_id=resource_id,
        action_value=1, # 임시로 1개 차감
        accumulated_tokens=accumulated_tokens,
        inference_latency_sec=inference_latency_sec
    )
    
    try:
        # Go 서버로 슛! 하고 결과 받기
        response = stub.CommitTransaction(request)
        
        # 에이전트에게 결과를 문자열로 반환 (에이전트는 이 결과를 보고 다음 행동을 결정함)
        if response.is_rolled_back:
            return f"실패: {response.message} (비용 방어액: ${response.saved_cost_usd})"
        else:
            return f"성공: {response.message}"
            
    except grpc.RpcError as e:
        return f"통신 에러 발생: {e.details()}"

# 3. 직접 실행해서 통신이 잘 되는지 테스트해보는 코드
if __name__ == "__main__":
    # Go 서버가 켜져 있어야만 작동해!
    print("--- 에이전트 -> 미들웨어 통신 테스트 ---")
    
    # 가상의 에이전트 1번이 1500토큰, 4.5초 대기 후 결제를 요청한다고 가정
    result = request_commit_tool(
        agent_id="Agent_001",
        resource_id="flight_ticket_A",
        accumulated_tokens=1500,
        inference_latency_sec=4.5
    )
    
    print(f"미들웨어 응답: {result}")