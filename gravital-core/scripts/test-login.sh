#!/bin/bash

# 测试登录接口

echo "🧪 测试 Gravital Core 登录接口..."
echo ""

# API 地址
API_URL=${API_URL:-http://localhost:8080}

# 登录信息
USERNAME=${USERNAME:-admin}
PASSWORD=${PASSWORD:-admin123}

echo "📝 登录信息:"
echo "   API: $API_URL"
echo "   用户名: $USERNAME"
echo "   密码: $PASSWORD"
echo ""

echo "🔄 发送登录请求..."
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

echo ""
echo "📨 响应结果:"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
echo ""

# 检查是否成功
if echo "$RESPONSE" | grep -q "token"; then
    echo "✅ 登录成功！"
    
    # 提取 token
    TOKEN=$(echo "$RESPONSE" | jq -r '.data.token' 2>/dev/null)
    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        echo ""
        echo "🔑 Token (前50个字符):"
        echo "   ${TOKEN:0:50}..."
        echo ""
        echo "💡 可以使用以下命令测试 API:"
        echo "   curl -H \"Authorization: Bearer $TOKEN\" $API_URL/api/v1/auth/me"
    fi
else
    echo "❌ 登录失败！"
    exit 1
fi

