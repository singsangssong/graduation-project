const fs = require("fs");
const path = require("path");
const pptxgen = require("pptxgenjs");

const OUT_DIR = path.join(__dirname, "..", "deliverables", "final-presentation");
const OUT_FILE = path.join(OUT_DIR, "agentic-middleware-final.pptx");

fs.mkdirSync(OUT_DIR, { recursive: true });

const pptx = new pptxgen();
pptx.layout = "LAYOUT_WIDE";
pptx.author = "Song Seonghun";
pptx.subject = "Graduation project final presentation";
pptx.title = "Agentic Middleware";
pptx.company = "Kyung Hee University";
pptx.lang = "ko-KR";
pptx.theme = {
  headFontFace: "Apple SD Gothic Neo",
  bodyFontFace: "Apple SD Gothic Neo",
  lang: "ko-KR",
};
pptx.defineLayout({ name: "LAYOUT_WIDE", width: 13.333, height: 7.5 });

const C = {
  bg: "F5F7FB",
  ink: "172033",
  muted: "667085",
  line: "D9E2EF",
  blue: "2563EB",
  blueDark: "194185",
  green: "079455",
  red: "D92D20",
  amber: "B54708",
  teal: "0F766E",
  slate: "344054",
  white: "FFFFFF",
  black: "101828",
  paleBlue: "EAF2FF",
  paleGreen: "EAF8EF",
  paleAmber: "FFF4E5",
  paleRed: "FFF0F0",
};

function addBg(slide) {
  slide.background = { color: C.bg };
  slide.addShape(pptx.ShapeType.rect, {
    x: 0,
    y: 0,
    w: 13.333,
    h: 0.18,
    fill: { color: C.blue },
    line: { color: C.blue },
  });
}

function addTitle(slide, title, subtitle) {
  slide.addText(title, {
    x: 0.55,
    y: 0.42,
    w: 9.9,
    h: 0.42,
    fontFace: "Apple SD Gothic Neo",
    fontSize: 22,
    bold: true,
    color: C.ink,
    margin: 0,
    breakLine: false,
  });
  if (subtitle) {
    slide.addText(subtitle, {
      x: 0.56,
      y: 0.92,
      w: 11.4,
      h: 0.25,
      fontSize: 10.5,
      color: C.muted,
      margin: 0,
    });
  }
}

function addFooter(slide, idx) {
  slide.addText("Agentic Middleware Graduation Project", {
    x: 0.55,
    y: 7.08,
    w: 5,
    h: 0.18,
    fontSize: 8.5,
    color: C.muted,
    margin: 0,
  });
  slide.addText(String(idx).padStart(2, "0"), {
    x: 12.25,
    y: 7.05,
    w: 0.65,
    h: 0.22,
    fontSize: 9,
    bold: true,
    color: C.muted,
    align: "right",
    margin: 0,
  });
}

function card(slide, x, y, w, h, opts = {}) {
  slide.addShape(pptx.ShapeType.roundRect, {
    x,
    y,
    w,
    h,
    rectRadius: 0.08,
    fill: { color: opts.fill || C.white },
    line: { color: opts.line || C.line, width: opts.lineWidth || 0.9 },
    shadow: opts.shadow ? { type: "outer", color: "CBD5E1", opacity: 0.18, blur: 1.5, angle: 45, distance: 1 } : undefined,
  });
}

function pill(slide, text, x, y, w, color, fill) {
  slide.addShape(pptx.ShapeType.roundRect, {
    x,
    y,
    w,
    h: 0.28,
    rectRadius: 0.14,
    fill: { color: fill },
    line: { color: fill },
  });
  slide.addText(text, {
    x: x + 0.08,
    y: y + 0.06,
    w: w - 0.16,
    h: 0.11,
    fontSize: 8.5,
    bold: true,
    color,
    align: "center",
    margin: 0,
  });
}

function bulletList(slide, items, x, y, w, h, opts = {}) {
  const runs = [];
  for (const item of items) {
    runs.push({
      text: item,
      options: {
        bullet: { type: "ul" },
        breakLine: true,
      },
    });
  }
  slide.addText(runs, {
    x,
    y,
    w,
    h,
    fontSize: opts.fontSize || 15,
    color: opts.color || C.ink,
    fit: "shrink",
    paraSpaceAfterPt: opts.paraSpaceAfterPt || 8,
    breakLine: false,
    margin: 0,
  });
}

function metric(slide, label, value, note, x, y, w, h, color = C.blue) {
  card(slide, x, y, w, h, { fill: C.white, shadow: true });
  slide.addText(label, { x: x + 0.18, y: y + 0.16, w: w - 0.36, h: 0.2, fontSize: 9.5, bold: true, color: C.muted, margin: 0 });
  slide.addText(value, { x: x + 0.18, y: y + 0.5, w: w - 0.36, h: 0.45, fontSize: 24, bold: true, color, margin: 0 });
  slide.addText(note, { x: x + 0.18, y: y + 1.02, w: w - 0.36, h: 0.25, fontSize: 8.7, color: C.muted, margin: 0, fit: "shrink" });
}

function node(slide, text, x, y, w, h, fill, color = C.ink) {
  card(slide, x, y, w, h, { fill, line: C.line });
  slide.addText(text, {
    x: x + 0.08,
    y: y + 0.18,
    w: w - 0.16,
    h: h - 0.24,
    fontSize: 12,
    bold: true,
    color,
    align: "center",
    valign: "mid",
    margin: 0,
    fit: "shrink",
  });
}

function arrow(slide, x1, y1, x2, y2, color = C.slate) {
  slide.addShape(pptx.ShapeType.line, {
    x: x1,
    y: y1,
    w: x2 - x1,
    h: y2 - y1,
    line: { color, width: 1.5, beginArrowType: "none", endArrowType: "triangle" },
  });
}

function addNotes(slide, notes) {
  if (slide.addNotes) slide.addNotes(notes);
}

let n = 1;

// 1
{
  const slide = pptx.addSlide();
  slide.background = { color: "EEF4FF" };
  slide.addShape(pptx.ShapeType.rect, { x: 0, y: 0, w: 13.333, h: 7.5, fill: { color: "EEF4FF" }, line: { color: "EEF4FF" } });
  slide.addShape(pptx.ShapeType.rect, { x: 0, y: 0, w: 13.333, h: 0.22, fill: { color: C.blue }, line: { color: C.blue } });
  pill(slide, "Graduation Project Final Evaluation", 0.7, 0.82, 2.75, C.blueDark, "D7E7FF");
  slide.addText("Agentic Middleware", { x: 0.7, y: 1.55, w: 10.7, h: 0.62, fontSize: 35, bold: true, color: C.ink, margin: 0 });
  slide.addText("다중 LLM 에이전트 트랜잭션을 위한 QCFuse + ATCC 기반 미들웨어", {
    x: 0.72,
    y: 2.32,
    w: 10.8,
    h: 0.45,
    fontSize: 19,
    color: C.slate,
    margin: 0,
  });
  card(slide, 0.72, 3.24, 4.05, 1.08, { fill: C.white, shadow: true });
  slide.addText("Problem", { x: 0.98, y: 3.45, w: 1.4, h: 0.18, fontSize: 10, bold: true, color: C.red, margin: 0 });
  slide.addText("LLM reasoning이 길어질수록\nDB lock과 rollback 비용이 증폭", { x: 0.98, y: 3.76, w: 3.45, h: 0.34, fontSize: 12.5, color: C.ink, margin: 0, fit: "shrink" });
  card(slide, 4.98, 3.24, 4.05, 1.08, { fill: C.white, shadow: true });
  slide.addText("Solution", { x: 5.24, y: 3.45, w: 1.4, h: 0.18, fontSize: 10, bold: true, color: C.green, margin: 0 });
  slide.addText("QCFuse-style read fusion +\nATCC-style cost-aware arbitration", { x: 5.24, y: 3.76, w: 3.45, h: 0.34, fontSize: 12.5, color: C.ink, margin: 0, fit: "shrink" });
  card(slide, 9.24, 3.24, 2.75, 1.08, { fill: C.white, shadow: true });
  slide.addText("Demo", { x: 9.5, y: 3.45, w: 1.4, h: 0.18, fontSize: 10, bold: true, color: C.blue, margin: 0 });
  slide.addText("Go gRPC middleware\n+ Python agents", { x: 9.5, y: 3.76, w: 2.2, h: 0.34, fontSize: 12.5, color: C.ink, margin: 0, fit: "shrink" });
  slide.addText("송성훈 · 경희대학교 컴퓨터공학과", { x: 0.72, y: 6.72, w: 5.8, h: 0.24, fontSize: 11.5, color: C.muted, margin: 0 });
  addNotes(slide, "오늘 발표의 핵심은 Agentic AI 환경에서 트랜잭션 문제가 왜 비용 문제로 확장되는지, 그리고 이를 미들웨어에서 어떻게 제어했는지입니다.");
}
n++;

// 2
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "발표 흐름", "문제 정의에서 데모까지, 평가자가 따라오기 쉬운 구조로 구성");
  const items = [
    ["01", "Agentic AI 시대의 workload 변화", "단일 호출에서 다중 sub-agent 동시 실행으로 이동"],
    ["02", "기존 트랜잭션 제어의 한계", "긴 reasoning, lock 대기, rollback 비용 증폭"],
    ["03", "관련 연구와 gap", "SagaLLM, ATCC, QCFuse, AIOS의 역할 정리"],
    ["04", "제안 아키텍처와 구현", "Python agent layer + Go middleware layer"],
    ["05", "실험/데모/한계", "지표 기반 결과와 최종평가 데모 흐름"],
  ];
  items.forEach((it, i) => {
    const y = 1.45 + i * 0.95;
    card(slide, 0.8, y, 11.6, 0.68, { fill: C.white, shadow: true });
    slide.addText(it[0], { x: 1.05, y: y + 0.18, w: 0.5, h: 0.18, fontSize: 10, bold: true, color: C.blue, margin: 0 });
    slide.addText(it[1], { x: 1.75, y: y + 0.14, w: 3.5, h: 0.22, fontSize: 14, bold: true, color: C.ink, margin: 0 });
    slide.addText(it[2], { x: 5.35, y: y + 0.17, w: 6.65, h: 0.2, fontSize: 11.5, color: C.muted, margin: 0 });
  });
  addFooter(slide, n);
}
n++;

// 3
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "연구 배경: Agentic AI는 동시 실행 구조로 이동한다", "하나의 LLM 응답이 아니라, 여러 agent가 계획·도구호출·검증을 병렬 수행하는 구조");
  node(slide, "User Goal", 0.75, 2.85, 1.35, 0.65, C.white);
  node(slide, "Planner\nLLM", 2.65, 2.7, 1.45, 0.95, C.paleBlue, C.blueDark);
  const agents = [
    ["Search\nAgent", 5.1, 1.35],
    ["DB\nAgent", 5.1, 2.62],
    ["Payment\nAgent", 5.1, 3.89],
    ["Verifier\nAgent", 8.35, 2.62],
  ];
  agents.forEach(([t, x, y]) => node(slide, t, x, y, 1.7, 0.82, C.white));
  node(slide, "Shared\nResources", 10.92, 2.52, 1.55, 1.02, C.paleAmber, C.amber);
  arrow(slide, 2.12, 3.18, 2.62, 3.18, C.blue);
  arrow(slide, 4.1, 3.15, 5.05, 1.75, C.blue);
  arrow(slide, 4.1, 3.15, 5.05, 3.05, C.blue);
  arrow(slide, 4.1, 3.15, 5.05, 4.31, C.blue);
  arrow(slide, 6.85, 1.76, 10.9, 2.95, C.amber);
  arrow(slide, 6.85, 3.04, 10.9, 3.03, C.amber);
  arrow(slide, 6.85, 4.32, 10.9, 3.12, C.amber);
  arrow(slide, 10.9, 3.03, 10.05, 3.03, C.teal);
  slide.addText("동시성은 agent의 성능을 높이지만,\n공유 자원 접근 시 transaction conflict를 만든다.", {
    x: 0.9,
    y: 5.7,
    w: 11.6,
    h: 0.55,
    fontSize: 18,
    bold: true,
    color: C.ink,
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 4
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "문제 정의: Agentic transaction은 기존 OLTP와 다르다", "짧고 예측 가능한 transaction이 아니라, 긴 reasoning과 비결정적 접근을 포함한다");
  const rows = [
    ["구분", "기존 OLTP transaction", "Agentic transaction"],
    ["실행 시간", "수 ms~수십 ms", "LLM reasoning으로 수 초~수 분"],
    ["접근 패턴", "쿼리/락 범위 예측 가능", "도구 호출 결과에 따라 동적으로 변화"],
    ["Abort 비용", "DB 작업 일부 손실", "Token/API/reasoning latency 전체 손실"],
    ["충돌 처리", "무작위 abort도 비용 작음", "고비용 agent abort 시 경제적 손실 큼"],
  ];
  const x = 0.75;
  const y = 1.45;
  const widths = [1.65, 4.35, 6.0];
  rows.forEach((row, i) => {
    const yy = y + i * 0.82;
    const fill = i === 0 ? C.blue : C.white;
    const color = i === 0 ? C.white : C.ink;
    let xx = x;
    row.forEach((cell, j) => {
      slide.addShape(pptx.ShapeType.rect, { x: xx, y: yy, w: widths[j], h: 0.72, fill: { color: fill }, line: { color: C.line, width: 0.8 } });
      slide.addText(cell, { x: xx + 0.12, y: yy + 0.19, w: widths[j] - 0.24, h: 0.18, fontSize: i === 0 ? 11 : 10.5, bold: i === 0 || j === 0, color, margin: 0, fit: "shrink" });
      xx += widths[j];
    });
  });
  card(slide, 0.85, 6.0, 11.7, 0.55, { fill: C.paleRed, line: "F5B5B2" });
  slide.addText("핵심 질문: 여러 agent가 같은 자원에 접근할 때, 시스템은 어떤 요청을 보호하고 어떤 요청을 rollback해야 하는가?", {
    x: 1.05,
    y: 6.18,
    w: 11.25,
    h: 0.18,
    fontSize: 13,
    bold: true,
    color: C.red,
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 5
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "관련 연구: transaction 보장, scheduling, cache fusion은 각각 분리되어 있다", "본 프로젝트는 세 흐름을 agentic middleware layer에서 결합한다");
  const works = [
    ["Concurrent Modular Agent", "다중 LLM 모듈의 비동기 병렬 실행", "공유 DB 충돌/rollback cost는 직접 해결하지 않음"],
    ["AIOS", "Agent OS 관점의 scheduler/tool/memory 관리", "DB transaction guard는 별도 구현 필요"],
    ["SagaLLM", "LLM workflow에 Saga compensation 도입", "동시 경합 시 비용 기반 rollback 우선순위는 부족"],
    ["ATCC", "Agentic transaction의 abort cost-aware CC", "본 프로젝트는 token/latency 기반 단순화 구현"],
    ["QCFuse", "Query-centric KV cache fusion", "본 프로젝트는 read request fusion으로 재해석"],
  ];
  works.forEach((w, i) => {
    const yy = 1.32 + i * 0.94;
    card(slide, 0.7, yy, 12.0, 0.72, { fill: i % 2 ? "FBFCFF" : C.white });
    slide.addText(w[0], { x: 0.95, y: yy + 0.19, w: 2.65, h: 0.18, fontSize: 10.5, bold: true, color: C.blueDark, margin: 0, fit: "shrink" });
    slide.addText(w[1], { x: 3.75, y: yy + 0.19, w: 3.65, h: 0.18, fontSize: 10.5, color: C.ink, margin: 0, fit: "shrink" });
    slide.addText(w[2], { x: 7.62, y: yy + 0.19, w: 4.55, h: 0.18, fontSize: 10.5, color: C.muted, margin: 0, fit: "shrink" });
  });
  pill(slide, "Gap: 다중 agent transaction에서 read amplification + lock wait + sunk-cost rollback을 함께 다루는 middleware", 1.1, 6.18, 11.1, C.white, C.blue);
  addFooter(slide, n);
}
n++;

// 6
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "제안 아키텍처: Python Agent 계층과 Go Middleware 계층 분리", "gRPC typed contract를 기준으로 agent runtime과 transaction control을 분리");
  node(slide, "Python Agent Layer\nAutoGen / LangGraph style\nTool Call + Stress Test", 0.8, 1.55, 2.75, 1.25, C.paleBlue, C.blueDark);
  node(slide, "gRPC + Protocol Buffers\nproto/middleware.proto", 4.05, 1.68, 2.5, 1.0, C.white, C.ink);
  node(slide, "Go Middleware Server\nQCFuse Scheduler\nATCC Scheduler\nMetrics API", 7.0, 1.42, 2.85, 1.52, C.paleGreen, C.green);
  node(slide, "Resource Layer\nTicket Stock Simulation\nFuture: MySQL / Redis", 10.45, 1.55, 2.15, 1.25, C.paleAmber, C.amber);
  arrow(slide, 3.55, 2.18, 4.0, 2.18, C.blue);
  arrow(slide, 6.55, 2.18, 7.0, 2.18, C.blue);
  arrow(slide, 9.85, 2.18, 10.45, 2.18, C.amber);
  const phases = [
    ["Phase 1", "ReadResource", "동일 자원 조회 요청을 짧은 window에 모아 fusion"],
    ["Phase 2", "Lock-Free Reasoning", "LLM reasoning 동안 DB connection/lock을 점유하지 않음"],
    ["Phase 3", "CommitTransaction", "token + latency 기반 sunk cost로 winner 선정"],
  ];
  phases.forEach((p, i) => {
    const xx = 0.92 + i * 4.15;
    card(slide, xx, 4.25, 3.55, 1.05, { fill: C.white, shadow: true });
    slide.addText(p[0], { x: xx + 0.2, y: 4.45, w: 0.9, h: 0.16, fontSize: 9, bold: true, color: C.blue, margin: 0 });
    slide.addText(p[1], { x: xx + 1.05, y: 4.43, w: 2.0, h: 0.18, fontSize: 12, bold: true, color: C.ink, margin: 0 });
    slide.addText(p[2], { x: xx + 0.2, y: 4.77, w: 3.1, h: 0.26, fontSize: 9.3, color: C.muted, margin: 0, fit: "shrink" });
  });
  addFooter(slide, n);
}
n++;

// 7
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "QCFuse-style Read Fusion: 중복 조회를 하나의 I/O로 압축", "원 논문의 KV cache fusion 아이디어를 middleware read path의 request fusion으로 재해석");
  const steps = [
    ["1", "Windowing", "100ms 안에 동일 resource_id 조회 요청을 수집"],
    ["2", "Fusion", "N개의 read request를 하나의 logical DB read로 병합"],
    ["3", "Broadcast", "단일 조회 결과를 대기 중인 모든 agent에게 반환"],
    ["4", "Metrics", "saved_db_reads = request_count - fused_batch_count"],
  ];
  steps.forEach((s, i) => {
    const x = 0.8 + i * 3.15;
    card(slide, x, 1.55, 2.65, 2.15, { fill: C.white, shadow: true });
    slide.addText(s[0], { x: x + 0.2, y: 1.78, w: 0.35, h: 0.22, fontSize: 13, bold: true, color: C.blue, margin: 0 });
    slide.addText(s[1], { x: x + 0.58, y: 1.78, w: 1.85, h: 0.22, fontSize: 13, bold: true, color: C.ink, margin: 0 });
    slide.addText(s[2], { x: x + 0.22, y: 2.28, w: 2.1, h: 0.56, fontSize: 11.2, color: C.muted, margin: 0, fit: "shrink" });
    if (i < steps.length - 1) arrow(slide, x + 2.65, 2.62, x + 3.1, 2.62, C.blue);
  });
  metric(slide, "Deterministic Demo", "5 → 1", "5개의 read를 1개 batch로 융합", 1.2, 4.7, 3.25, 1.35, C.green);
  metric(slide, "Saved DB Reads", "4", "동일 자원 조회에서 줄인 I/O", 4.95, 4.7, 3.25, 1.35, C.blue);
  metric(slide, "Stress Target", "10/50/100/200", "최종보고서용 확장 실험", 8.7, 4.7, 3.25, 1.35, C.amber);
  addFooter(slide, n);
}
n++;

// 8
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "Lock-Free Reasoning: LLM 추론 시간은 DB lock으로 전이되지 않아야 한다", "Agent는 조회 후 독립적으로 reasoning하고, commit 시점에만 middleware arbitration에 참여");
  node(slide, "ReadResource\n짧은 조회", 0.9, 2.1, 1.75, 0.85, C.paleBlue, C.blueDark);
  node(slide, "Release DB/Resource\nNo Lock Held", 3.2, 2.1, 2.05, 0.85, C.paleGreen, C.green);
  node(slide, "LLM Reasoning\n2~10 sec", 5.9, 2.1, 2.0, 0.85, C.white, C.ink);
  node(slide, "CommitTransaction\nATCC Queue", 8.55, 2.1, 2.1, 0.85, C.paleAmber, C.amber);
  node(slide, "Commit / Rollback\nSignal", 11.0, 2.1, 1.7, 0.85, C.white, C.ink);
  arrow(slide, 2.65, 2.53, 3.2, 2.53, C.blue);
  arrow(slide, 5.25, 2.53, 5.9, 2.53, C.green);
  arrow(slide, 7.9, 2.53, 8.55, 2.53, C.amber);
  arrow(slide, 10.65, 2.53, 11.0, 2.53, C.amber);
  card(slide, 1.15, 4.45, 11.0, 1.08, { fill: C.white, shadow: true });
  slide.addText("발표 메시지", { x: 1.45, y: 4.72, w: 1.2, h: 0.18, fontSize: 10, bold: true, color: C.blue, margin: 0 });
  slide.addText("LLM이 생각하는 동안 DB lock을 붙잡지 않는다. 긴 reasoning은 agent layer에서 진행하고, conflict resolution은 commit phase에서만 처리한다.", {
    x: 2.72,
    y: 4.7,
    w: 8.9,
    h: 0.24,
    fontSize: 13,
    bold: true,
    color: C.ink,
    margin: 0,
    fit: "shrink",
  });
  addFooter(slide, n);
}
n++;

// 9
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "ATCC-style Cost-Aware Arbitration: 매몰 비용이 큰 agent를 보호", "원 ATCC의 abort-cost-aware priority scheduling을 token/latency 기반 score로 단순화");
  card(slide, 0.85, 1.45, 5.35, 1.35, { fill: C.white, shadow: true });
  slide.addText("Sunk Cost Score", { x: 1.15, y: 1.72, w: 2.4, h: 0.25, fontSize: 15, bold: true, color: C.blueDark, margin: 0 });
  slide.addText("cost = tokens × 0.002 + latency_sec × 0.5", { x: 1.15, y: 2.12, w: 4.7, h: 0.25, fontSize: 15, bold: true, color: C.ink, margin: 0 });
  card(slide, 6.65, 1.45, 5.75, 1.35, { fill: C.paleAmber, line: "FAD7A0" });
  slide.addText("Commit Rule", { x: 6.95, y: 1.72, w: 2.4, h: 0.25, fontSize: 15, bold: true, color: C.amber, margin: 0 });
  slide.addText("경합 batch를 score 내림차순 정렬 → 1등 commit, 나머지는 rollback signal", { x: 6.95, y: 2.12, w: 5.0, h: 0.25, fontSize: 13.5, bold: true, color: C.ink, margin: 0, fit: "shrink" });
  const agents = [
    ["Agent-A", "$13.25", "WIN", C.green],
    ["Agent-D", "$8.75", "ROLLBACK", C.amber],
    ["Agent-C", "$5.10", "ROLLBACK", C.amber],
    ["Agent-E", "$2.85", "ROLLBACK", C.amber],
    ["Agent-B", "$1.00", "ROLLBACK", C.amber],
  ];
  agents.forEach((a, i) => {
    const yy = 3.35 + i * 0.48;
    card(slide, 2.0, yy, 9.25, 0.36, { fill: i === 0 ? C.paleGreen : C.white, line: i === 0 ? "A9E8C4" : C.line });
    slide.addText(String(i + 1), { x: 2.22, y: yy + 0.11, w: 0.25, h: 0.1, fontSize: 8.5, bold: true, color: C.muted, margin: 0 });
    slide.addText(a[0], { x: 2.65, y: yy + 0.09, w: 1.4, h: 0.12, fontSize: 10.5, bold: true, color: C.ink, margin: 0 });
    slide.addText(a[1], { x: 5.0, y: yy + 0.09, w: 1.0, h: 0.12, fontSize: 10.5, bold: true, color: a[3], margin: 0 });
    slide.addText(a[2], { x: 8.9, y: yy + 0.09, w: 1.4, h: 0.12, fontSize: 9.5, bold: true, color: a[3], align: "right", margin: 0 });
  });
  addFooter(slide, n);
}
n++;

// 10
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "시스템 구현: 최소하지만 검증 가능한 middleware prototype", "Go channel scheduler, gRPC contract, Python demo runner, metrics dashboard");
  const blocks = [
    ["Go Middleware", "QCFuseScheduler / ATCCScheduler\nmetrics endpoint / config env"],
    ["Protocol Buffers", "ReadResource / CommitTransaction\nPython-Go typed contract"],
    ["Python Agent", "deterministic_demo.py\nstress_test_v2.py / AutoGen scenario"],
    ["Dashboard", "live metrics / leaderboard\nvideo recording script"],
  ];
  blocks.forEach((b, i) => {
    const x = 0.9 + (i % 2) * 6.2;
    const y = 1.55 + Math.floor(i / 2) * 2.05;
    card(slide, x, y, 5.55, 1.48, { fill: C.white, shadow: true });
    slide.addText(b[0], { x: x + 0.25, y: y + 0.25, w: 2.6, h: 0.22, fontSize: 15, bold: true, color: C.blueDark, margin: 0 });
    slide.addText(b[1], { x: x + 0.25, y: y + 0.72, w: 4.95, h: 0.42, fontSize: 11.5, color: C.muted, margin: 0, fit: "shrink" });
  });
  card(slide, 1.1, 6.1, 11.1, 0.46, { fill: C.paleBlue, line: "C7DDFF" });
  slide.addText("핵심 원칙: agent framework 내부가 아니라 middleware layer에서 공통 transaction guard를 제공한다.", {
    x: 1.35,
    y: 6.24,
    w: 10.6,
    h: 0.14,
    fontSize: 12,
    bold: true,
    color: C.blueDark,
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 11
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "LLM Tool Call 안정화: hallucination을 실행 가능한 contract로 제한", "AutoGen/LangGraph 계층에서 발생할 수 있는 파라미터 누락·타입 오류·도구 미호출 문제 대응");
  const problems = [
    ["문제 1", "LLM이 token_cost를 문자열로 보내거나 필드를 누락"],
    ["문제 2", "도구를 호출하지 않고 자연어로만 '결제 완료'라고 응답"],
    ["문제 3", "middleware 통신 에러가 agent workflow 전체 crash로 이어짐"],
  ];
  problems.forEach((p, i) => {
    card(slide, 0.9, 1.55 + i * 0.88, 5.3, 0.62, { fill: C.paleRed, line: "F5B5B2" });
    slide.addText(p[0], { x: 1.15, y: 1.76 + i * 0.88, w: 0.8, h: 0.14, fontSize: 9, bold: true, color: C.red, margin: 0 });
    slide.addText(p[1], { x: 2.0, y: 1.73 + i * 0.88, w: 3.95, h: 0.18, fontSize: 10.5, color: C.ink, margin: 0, fit: "shrink" });
  });
  const fixes = [
    ["Type Hint + Docstring", "tool schema를 명확히 하여 LLM이 올바른 JSON argument를 구성하도록 유도"],
    ["Canonical Proto", "Python/Go 사이 필드명과 타입을 하나의 proto 계약으로 고정"],
    ["Self-healing Message", "통신 에러를 문자열로 반환해 agent가 재시도/수정할 수 있게 함"],
  ];
  fixes.forEach((f, i) => {
    card(slide, 6.8, 1.55 + i * 0.88, 5.4, 0.62, { fill: C.paleGreen, line: "A9E8C4" });
    slide.addText(f[0], { x: 7.05, y: 1.76 + i * 0.88, w: 1.8, h: 0.14, fontSize: 9, bold: true, color: C.green, margin: 0 });
    slide.addText(f[1], { x: 8.75, y: 1.73 + i * 0.88, w: 3.1, h: 0.18, fontSize: 10.2, color: C.ink, margin: 0, fit: "shrink" });
  });
  node(slide, "LLM", 1.55, 5.35, 1.2, 0.5, C.white);
  node(slide, "Tool Schema", 3.35, 5.35, 1.7, 0.5, C.paleBlue, C.blueDark);
  node(slide, "gRPC Proto", 5.75, 5.35, 1.65, 0.5, C.white);
  node(slide, "Go Middleware", 8.15, 5.35, 1.9, 0.5, C.paleGreen, C.green);
  node(slide, "Result/Error", 10.75, 5.35, 1.55, 0.5, C.white);
  arrow(slide, 2.75, 5.6, 3.35, 5.6, C.blue);
  arrow(slide, 5.05, 5.6, 5.75, 5.6, C.blue);
  arrow(slide, 7.4, 5.6, 8.15, 5.6, C.blue);
  arrow(slide, 10.05, 5.6, 10.75, 5.6, C.teal);
  addFooter(slide, n);
}
n++;

// 12
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "실험 설계: 동일 자원 재고 1장에 대한 다중 agent 경합", "주차 발표에서 사용한 10/50/100/200 agent stress 시나리오를 최종평가용으로 정리");
  metric(slide, "Resource", "1 ticket", "flight_ticket_A", 0.9, 1.55, 2.7, 1.35, C.blue);
  metric(slide, "Agents", "10~200", "동시 접속 thread", 3.9, 1.55, 2.7, 1.35, C.green);
  metric(slide, "Reasoning", "2~10s", "LLM latency simulation", 6.9, 1.55, 2.7, 1.35, C.amber);
  metric(slide, "Cost", "$ proxy", "token + latency", 9.9, 1.55, 2.7, 1.35, C.red);
  const flow = [
    "동시 진입: 모든 agent가 동일 resource 조회",
    "QCFuse: 짧은 window에서 read request fusion",
    "Lock-Free: agent별 reasoning latency simulation",
    "ATCC: commit batch leaderboard 정렬",
    "결과: winner commit, loser rollback signal",
  ];
  bulletList(slide, flow, 1.2, 3.65, 11.0, 1.5, { fontSize: 15 });
  card(slide, 1.2, 6.0, 11.0, 0.52, { fill: C.white, shadow: true });
  slide.addText("측정 지표: saved DB reads, commit/rollback 수, total saved cost, elapsed time, throughput", {
    x: 1.55,
    y: 6.18,
    w: 10.3,
    h: 0.16,
    fontSize: 12,
    bold: true,
    color: C.ink,
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 13
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "Deterministic Demo Result: 발표 중 항상 같은 결과를 재현", "5명의 agent를 고정 비용으로 실행하여 dashboard와 JSON/CSV 결과를 함께 보여준다");
  metric(slide, "Winner", "Agent-A", "highest sunk cost", 0.8, 1.45, 2.6, 1.4, C.green);
  metric(slide, "Rollbacks", "4", "Saga-style rollback", 3.65, 1.45, 2.6, 1.4, C.amber);
  metric(slide, "Saved DB Reads", "4", "QCFuse read fusion", 6.5, 1.45, 2.6, 1.4, C.blue);
  metric(slide, "Saved Cost", "$17.70", "ATCC rollback value", 9.35, 1.45, 2.6, 1.4, C.red);
  const code = [
    "cd middleware-go && GOCACHE=/private/tmp/agenic-middleware-gocache go run .",
    "cd agent-python && python3 deterministic_demo.py",
    "open dashboard.html",
  ];
  card(slide, 0.9, 3.65, 11.5, 1.45, { fill: C.black, line: C.black });
  slide.addText(code.join("\n"), { x: 1.25, y: 4.0, w: 10.9, h: 0.6, fontFace: "Menlo", fontSize: 12, color: "D0D5DD", margin: 0, fit: "shrink" });
  card(slide, 1.15, 5.78, 10.95, 0.5, { fill: C.paleBlue, line: "C7DDFF" });
  slide.addText("발표 포인트: 단순 성공/실패가 아니라, middleware가 절약한 I/O와 보호한 비용을 숫자로 보여준다.", {
    x: 1.45,
    y: 5.95,
    w: 10.4,
    h: 0.15,
    fontSize: 12,
    bold: true,
    color: C.blueDark,
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 14
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "Stress Test 관찰 결과: agent 수가 늘어도 rollback/error는 제어 가능", "12주차 자료의 10/50/100/200 thread 실험 결과를 최종평가용으로 재정리");
  const rows = [
    ["Agents", "처리 시간", "Success", "Rollback", "Error"],
    ["10", "~10 sec", "3", "7", "0"],
    ["50", "~10 sec", "4", "46", "0"],
    ["100", "~10 sec", "4", "96", "0"],
    ["200", "~10 sec", "4", "196", "0"],
  ];
  const x = 1.05;
  const y = 1.45;
  const widths = [1.65, 2.45, 2.25, 2.25, 2.25];
  rows.forEach((row, i) => {
    const yy = y + i * 0.68;
    const fill = i === 0 ? C.blue : C.white;
    const color = i === 0 ? C.white : C.ink;
    let xx = x;
    row.forEach((cell, j) => {
      slide.addShape(pptx.ShapeType.rect, { x: xx, y: yy, w: widths[j], h: 0.58, fill: { color: fill }, line: { color: C.line, width: 0.8 } });
      slide.addText(cell, { x: xx + 0.08, y: yy + 0.17, w: widths[j] - 0.16, h: 0.13, fontSize: 10.5, bold: i === 0, color, align: "center", margin: 0 });
      xx += widths[j];
    });
  });
  card(slide, 1.1, 5.25, 11.1, 0.88, { fill: C.white, shadow: true });
  slide.addText("해석", { x: 1.38, y: 5.52, w: 0.9, h: 0.16, fontSize: 10, bold: true, color: C.blue, margin: 0 });
  slide.addText("재고 1장이지만 ATCC window 단위로 batch가 반복되며 각 window마다 1건 승인된다. 최종보고서에서는 window 설정과 재고 reset 정책을 명확히 설명한다.", {
    x: 2.3,
    y: 5.49,
    w: 9.35,
    h: 0.23,
    fontSize: 11.5,
    color: C.ink,
    margin: 0,
    fit: "shrink",
  });
  addFooter(slide, n);
}
n++;

// 15
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "경제적 효과: 무작위 rollback보다 고비용 agent를 보호한다", "12주차 50명 경합 분석의 핵심 메시지를 비용 관점으로 정리");
  metric(slide, "Random-like Loss", "$405.74", "무작위/최악 승인 시 손실 예시", 0.9, 1.5, 3.55, 1.42, C.red);
  metric(slide, "ATCC Loss", "$374.46", "비용 기반 승인 적용", 4.75, 1.5, 3.55, 1.42, C.green);
  metric(slide, "Protected Value", "$31.28", "약 4.2만원 상당", 8.6, 1.5, 3.55, 1.42, C.blue);
  card(slide, 1.0, 3.75, 11.3, 1.25, { fill: C.white, shadow: true });
  slide.addText("핵심 주장", { x: 1.35, y: 4.08, w: 1.3, h: 0.2, fontSize: 12, bold: true, color: C.blue, margin: 0 });
  slide.addText("Agentic workload에서 rollback은 단순 DB 작업 취소가 아니라 이미 소모한 LLM token/API/reasoning time의 폐기다. 따라서 conflict resolution은 비용을 인지해야 한다.", {
    x: 2.7,
    y: 4.04,
    w: 8.9,
    h: 0.36,
    fontSize: 14,
    bold: true,
    color: C.ink,
    margin: 0,
    fit: "shrink",
  });
  card(slide, 2.0, 5.72, 9.4, 0.45, { fill: C.paleAmber, line: "FAD7A0" });
  slide.addText("주의: 최종보고서에서는 simulation 조건과 cost proxy임을 명확히 표시", { x: 2.35, y: 5.88, w: 8.7, h: 0.12, fontSize: 10.5, bold: true, color: C.amber, align: "center", margin: 0 });
  addFooter(slide, n);
}
n++;

// 16
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "데모 시연: Dashboard에서 middleware의 결정 과정을 실시간으로 확인", "교수님께 보여줄 영상/실시간 시연 흐름");
  const steps = [
    ["1", "Go middleware 실행", "gRPC :50051 / metrics :8080"],
    ["2", "dashboard.html 열기", "live 상태와 metric 초기값 확인"],
    ["3", "deterministic_demo.py 실행", "5 agents read + reasoning + commit"],
    ["4", "결과 설명", "Agent-A winner, rollback 4, saved reads 4, saved cost $17.70"],
  ];
  steps.forEach((s, i) => {
    const x = 0.85 + i * 3.1;
    card(slide, x, 1.55, 2.62, 2.1, { fill: C.white, shadow: true });
    slide.addText(s[0], { x: x + 0.22, y: 1.78, w: 0.35, h: 0.22, fontSize: 14, bold: true, color: C.blue, margin: 0 });
    slide.addText(s[1], { x: x + 0.22, y: 2.25, w: 2.1, h: 0.24, fontSize: 13, bold: true, color: C.ink, margin: 0, fit: "shrink" });
    slide.addText(s[2], { x: x + 0.22, y: 2.82, w: 2.05, h: 0.34, fontSize: 9.8, color: C.muted, margin: 0, fit: "shrink" });
  });
  card(slide, 1.25, 5.05, 10.85, 0.82, { fill: C.black, line: C.black });
  slide.addText("영상 포인트: terminal command → dashboard metric 변화 → JSON/CSV 결과 파일 확인", {
    x: 1.55,
    y: 5.36,
    w: 10.2,
    h: 0.14,
    fontSize: 12.5,
    bold: true,
    color: "D0D5DD",
    align: "center",
    margin: 0,
  });
  addFooter(slide, n);
}
n++;

// 17
{
  const slide = pptx.addSlide();
  addBg(slide);
  addTitle(slide, "프로젝트 기여와 한계", "완성된 prototype의 의미와 최종평가 이후 확장 방향");
  const contributions = [
    "Agentic transaction 문제를 token/API/reasoning cost 관점으로 재정의",
    "QCFuse-style read fusion과 ATCC-style arbitration을 하나의 middleware로 결합",
    "Python agent layer와 Go middleware layer를 gRPC contract로 분리",
    "Metrics API와 dashboard로 middleware 결정 과정을 시각화",
  ];
  const limits = [
    "현재 resource layer는 DB simulation이며 실제 MySQL/Redis 연동은 향후 과제",
    "Cost formula는 token/latency proxy 기반의 단순화 모델",
    "ATCC 원 논문의 RL 기반 phase-aware policy는 구현 범위 밖",
    "AutoGen/LangGraph full workflow는 optional scenario로 추가 안정화 필요",
  ];
  card(slide, 0.85, 1.45, 5.95, 4.75, { fill: C.white, shadow: true });
  slide.addText("기여", { x: 1.2, y: 1.75, w: 1.0, h: 0.24, fontSize: 16, bold: true, color: C.green, margin: 0 });
  bulletList(slide, contributions, 1.25, 2.25, 5.2, 2.8, { fontSize: 12.6 });
  card(slide, 7.1, 1.45, 5.45, 4.75, { fill: C.white, shadow: true });
  slide.addText("한계와 향후 계획", { x: 7.45, y: 1.75, w: 2.2, h: 0.24, fontSize: 16, bold: true, color: C.amber, margin: 0 });
  bulletList(slide, limits, 7.5, 2.25, 4.75, 2.8, { fontSize: 12.6 });
  addFooter(slide, n);
}
n++;

// 18
{
  const slide = pptx.addSlide();
  slide.background = { color: "EEF4FF" };
  slide.addShape(pptx.ShapeType.rect, { x: 0, y: 0, w: 13.333, h: 0.22, fill: { color: C.blue }, line: { color: C.blue } });
  slide.addText("결론", { x: 0.75, y: 0.95, w: 2.0, h: 0.32, fontSize: 20, bold: true, color: C.blueDark, margin: 0 });
  slide.addText("Agentic AI 환경에서 트랜잭션 제어는\n정합성뿐 아니라 비용 보존의 문제가 된다.", {
    x: 0.75,
    y: 1.55,
    w: 11.5,
    h: 1.15,
    fontSize: 29,
    bold: true,
    color: C.ink,
    margin: 0,
    fit: "shrink",
  });
  card(slide, 0.9, 3.55, 3.8, 1.2, { fill: C.white, shadow: true });
  slide.addText("QCFuse-style", { x: 1.2, y: 3.9, w: 2.2, h: 0.24, fontSize: 16, bold: true, color: C.blue, margin: 0 });
  slide.addText("중복 read I/O를 fusion", { x: 1.2, y: 4.28, w: 2.8, h: 0.18, fontSize: 11.5, color: C.muted, margin: 0 });
  card(slide, 4.9, 3.55, 3.8, 1.2, { fill: C.white, shadow: true });
  slide.addText("Lock-Free", { x: 5.2, y: 3.9, w: 2.2, h: 0.24, fontSize: 16, bold: true, color: C.green, margin: 0 });
  slide.addText("reasoning 동안 DB lock 해제", { x: 5.2, y: 4.28, w: 2.8, h: 0.18, fontSize: 11.5, color: C.muted, margin: 0 });
  card(slide, 8.9, 3.55, 3.4, 1.2, { fill: C.white, shadow: true });
  slide.addText("ATCC-style", { x: 9.2, y: 3.9, w: 2.2, h: 0.24, fontSize: 16, bold: true, color: C.amber, margin: 0 });
  slide.addText("매몰 비용 기반 commit arbitration", { x: 9.2, y: 4.28, w: 2.6, h: 0.18, fontSize: 11.5, color: C.muted, margin: 0, fit: "shrink" });
  slide.addText("Q & A", { x: 0.75, y: 6.35, w: 2.0, h: 0.32, fontSize: 18, bold: true, color: C.blueDark, margin: 0 });
  slide.addText("GitHub: https://github.com/singsangssong/graduation-project", { x: 6.9, y: 6.43, w: 5.6, h: 0.18, fontSize: 10.5, color: C.muted, align: "right", margin: 0 });
}

pptx.writeFile({ fileName: OUT_FILE });
console.log(OUT_FILE);
