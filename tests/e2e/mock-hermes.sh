#!/usr/bin/env bash
# Mock HermesAgent CLI：从 stdin 读 prompt，单行 echo 回去。供 e2e + adapter 测试使用。
read -r prompt
echo "Hermes echo: $prompt"
