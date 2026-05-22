#!/usr/bin/env bash
# Mock OpenClaw CLI：从 stdin 读一行 prompt，逐字流式输出。
# E2E 测试用这个脚本顶替真实 openclaw。
read -r prompt
echo "received prompt: $prompt"
for word in Hello from mock OpenClaw, you said: $prompt; do
    echo "$word"
    sleep 0.05
done
