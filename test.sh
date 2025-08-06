#!/bin/bash
base_url="http://localhost:8080"

echo "🔄 Создаём новую задачу..."

response=$(curl -s -w "\n%{http_code}" -X GET "$base_url/task")

body=$(echo "$response" | sed '$d')
code=$(echo "$response" | tail -n1)

if [ "$code" -ne 200 ]; then
    echo "❌ Ошибка при создании задачи: HTTP $code"
    echo "$body"
    exit 1
fi

task=$(echo "$body" | jq -r '.task_id // .taskId // empty')

if [ -z "$task" ] || [ "$task" = "null" ]; then
    echo "❌ Не удалось извлечь taskID из ответа:"
    echo "$body" | jq '.' 2>/dev/null || echo "$body"
    exit 1
fi

echo "✅ Создана задача: $task"

urls=(
    "https://img.freepik.com/free-photo/musk-duck-biziura-lobata-illustrated-by-elizabeth-gould_53876-65570.jpg"
    "https://img.freepik.com/free-photo/cook-s-petrel-procellaria-cookii-illustrated-by-elizabeth-gould_53876-65574.jpg"
    "https://img.freepik.com/free-photo/black-swan-cygnus-atratus-illustrated-by-elizabeth-gould_53876-65218.jpg"
)

base_url="http://localhost:8080"

echo "🎯 Task ID: $task"
echo "📤 Отправляем $(( ${#urls[@]} )) URL по одному..."

for url in "${urls[@]}"; do
    json_data=$(jq -n --arg u "$url" '{"url": $u}')

    echo "  ➤ Отправляем: $url"

    response=$(curl -s -w "\n%{http_code}" -X POST "$base_url/task/$task" \
      -H "Content-Type: application/json" \
      -d "$json_data")

    body=$(echo "$response" | sed '$d')
    code=$(echo "$response" | tail -n1)

    if [ "$code" -eq 200 ] || [ "$code" -eq 201 ]; then
        echo "    ✅ Успешно (HTTP $code)"
    else
        echo "    ❌ Ошибка (HTTP $code):"
        echo "$body" | jq -r '.' 2>/dev/null || echo "    $body"
    fi
done

echo "✅ Все URL отправлены."

echo "🔁 Проверяем статус задачи..."
while true; do
    status=$(curl -s -X GET "$base_url/task/$task" | jq -r '.status' 2>/dev/null || echo "unknown")

    case "$status" in
        "completed")
            echo "✅ Задача завершена!"
            break
            ;;
        "pending"|"processing")
            echo "⏳ Статус: $status, ждём..."
            sleep 2
            ;;
        "failed")
            echo "❌ Задача провалена."
            exit 1
            ;;
        *)
            echo "⚠️ Неизвестный статус: $status"
            sleep 2
            ;;
    esac
done

echo "⬇️ Скачиваем архив..."
curl -o "download_$task.zip" -O "$base_url/download/$task"

if [ -f "download_$task.zip" ] && [ ! -s "download_$task.zip" ]; then
    echo "❌ Файл пустой — возможно, ошибка при генерации."
    exit 1
else
    echo "📦 Успешно сохранено: download_$task.zip"
fi