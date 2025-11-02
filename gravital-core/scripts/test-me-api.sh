#!/bin/bash

# 测试 /api/v1/auth/me 接口

BASE_URL="http://localhost:8080"

echo "🧪 测试 /api/v1/auth/me 接口"
echo "================================"

# 1. 登录获取 token
echo ""
echo "📝 步骤 1: 登录获取 token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo "登录响应:"
echo "$LOGIN_RESPONSE" | jq '.'

# 提取 token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ 登录失败，无法获取 token"
  exit 1
fi

echo "✅ Token: ${TOKEN:0:20}..."

# 2. 测试 /api/v1/auth/me
echo ""
echo "📝 步骤 2: 测试 /api/v1/auth/me..."
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN")

echo "响应:"
echo "$ME_RESPONSE" | jq '.'

# 检查响应
CODE=$(echo "$ME_RESPONSE" | jq -r '.code')
if [ "$CODE" == "0" ]; then
  echo ""
  echo "✅ /api/v1/auth/me 接口测试成功！"
  echo ""
  echo "用户信息:"
  echo "$ME_RESPONSE" | jq '.data | {id, username, email, role: .role.name, permissions: .role.permissions}'
else
  echo ""
  echo "❌ /api/v1/auth/me 接口测试失败"
  echo "错误信息: $(echo "$ME_RESPONSE" | jq -r '.message')"
  exit 1
fi

echo ""
echo "================================"
echo "🎉 所有测试完成！"

