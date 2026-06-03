import asyncio
import os
import sys
from pathlib import Path

import grpc
from dotenv import load_dotenv

# AutoGen 패키지 임포트
from autogen_agentchat.agents import AssistantAgent
from autogen_agentchat.ui import Console
from autogen_ext.models.openai import OpenAIChatCompletionClient

# gRPC 컴파일된 파일은 proto/middleware.proto에서 생성된 파일만 사용합니다.
PROJECT_ROOT = Path(__file__).resolve().parent
AGENT_PYTHON_DIR = PROJECT_ROOT / "agent-python"
sys.path.insert(0, str(AGENT_PYTHON_DIR))

import middleware_pb2 as pb2
import middleware_pb2_grpc as pb2_grpc

# .env 파일에서 OpenAI API 키 로드
load_dotenv()

# ==========================================
# [파트 1] 미들웨어를 호출하는 도구(Tool) 정의
# ==========================================
async def purchase_ticket_via_middleware(agent_id: str, token_cost: float, latency: float) -> str:
    """항공권 결제를 요청하는 도구입니다. 이 도구를 사용하여 결제를 진행하세요."""
    print(f"\n[Tool 실행] {agent_id}가 미들웨어로 결제 요청을 전송합니다...")
    try:
        # Go 미들웨어 gRPC 서버(50051 포트)로 연결
        channel = grpc.aio.insecure_channel('localhost:50051')
        stub = pb2_grpc.TransactionMiddlewareStub(channel)
        
        request = pb2.CommitRequest(
            agent_id=agent_id,
            resource_id="flight_ticket_A",
            action_value=1,
            accumulated_tokens=int(token_cost),
            inference_latency_sec=latency
        )
        
        # 미들웨어에 요청 전송 및 응답 대기
        response = await stub.CommitTransaction(request)
        
        if response.is_rolled_back:
            return f"결제 실패 (Saga 롤백 발동). 사유: {response.message}"
        else:
            return f"결제 성공! 메시지: {response.message}"
    except Exception as e:
        return f"미들웨어 통신 오류: {str(e)}"

# ==========================================
# [파트 2] 에이전트 생성 도우미 함수
# ==========================================
def create_agent(agent_name: str, system_message: str) -> AssistantAgent:
    # LLM 클라이언트 설정 (GPT-4o-mini 사용)
    model_client = OpenAIChatCompletionClient(
        model="gpt-4o-mini",
        api_key=os.getenv("OPENAI_API_KEY")
    )
    
    # 에이전트 생성 및 도구(Tool) 장착
    agent = AssistantAgent(
        name=agent_name,
        system_message=system_message,
        model_client=model_client,
        tools=[purchase_ticket_via_middleware], # 위에서 만든 도구 장착!
        model_client_stream=True,
    )
    return agent

# ==========================================
# [파트 3] 개별 에이전트 작업 실행 함수
# ==========================================
async def run_agent_task(agent: AssistantAgent, prompt: str):
    # Console.run_stream을 통해 터미널에 에이전트의 생각 과정을 예쁘게 출력
    await Console(agent.run_stream(task=prompt))

# ==========================================
# [파트 4] 메인 시나리오 (동시성 경합 테스트)
# ==========================================
async def main():
    print("🚀 [시뮬레이션 시작] 다중 에이전트 동시 결제 경합 테스트 (티켓 1장)\n")
    
    # Agent A: 많은 비용(토큰)을 소모한 에이전트
    agent_a = create_agent(
        "Agent_A", 
        "당신은 항공권 예매 AI입니다. 결제 도구를 쓸 때 agent_id='Agent-A', token_cost=4500.0, latency=8.5 로 고정해서 호출하세요."
    )
    
    # Agent B: 적은 비용을 소모한 에이전트
    agent_b = create_agent(
        "Agent_B", 
        "당신은 항공권 예매 AI입니다. 결제 도구를 쓸 때 agent_id='Agent-B', token_cost=200.0, latency=1.2 로 고정해서 호출하세요."
    )
    
    task_prompt = "지금 당장 항공권 결제 도구를 사용해서 티켓을 1장 구매하고, 그 결과를 나에게 한국어로 보고해줘."
    
    # ★ asyncio.gather를 통해 두 에이전트에게 '동시에' 지시를 내림
    await asyncio.gather(
        run_agent_task(agent_a, task_prompt),
        run_agent_task(agent_b, task_prompt)
    )
    
    print("\n✅ [시뮬레이션 종료] 워크플로우 완료")

if __name__ == "__main__":
    asyncio.run(main())
