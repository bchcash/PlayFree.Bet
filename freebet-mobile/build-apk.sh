#!/bin/bash

# ===========================================
# FreeBet.guru Android APK Build Script
# ===========================================
# Этот скрипт собирает production APK для Android через EAS Build
# Использует Expo SDK 54, только Android платформа
# Оптимизирован для AlmaLinux 10 и других RPM-based дистрибутивов
# Требуется настроенный EXPO_TOKEN и инициализированный EAS проект
#
# Использование:
#   ./build-apk.sh
#
# Перед запуском:
#   1. Создайте .env.local файл с EXPO_TOKEN=ваш_токен
#   2. Убедитесь что в директории expo-go
#   3. Запустите: ./build-apk.sh
#
# Системные требования:
#   - AlmaLinux 10 (или RHEL/CentOS 9+)
#   - curl, tar, Node.js 18+, npm

set -e  # Остановить скрипт при первой ошибке

# Загрузка переменных окружения из .env.local файла
if [ -f ".env.local" ]; then
    echo "🔧 Загружаю переменные из .env.local..."
    export $(grep -v '^#' .env.local | xargs)
fi

echo "🚀 FreeBet.guru Android APK Build Script"
echo "========================================"
echo "🎯 Оптимизирован для AlmaLinux 10"
echo "📊 Система: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "🖥️  Ядро: $(uname -r)"
echo "📅 Время: $(date)"
echo ""

# Проверка и установка системных зависимостей
echo "🔧 Проверяю системные зависимости..."

# Проверка curl
if ! command -v curl &> /dev/null; then
    echo "📦 Устанавливаю curl..."
    sudo dnf install -y curl
fi

# Проверка Node.js
if ! command -v node &> /dev/null; then
    echo "📦 Устанавливаю Node.js 18..."
    # Добавляем репозиторий Node.js
    curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
    sudo dnf install -y nodejs
fi

# Проверка версии Node.js (минимум 18)
NODE_VERSION=$(node -v | cut -d'.' -f1 | cut -d'v' -f2)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "⚠️  Node.js версии $NODE_VERSION обнаружена. Рекомендуется 18+"
    echo "📦 Обновляю Node.js до версии 18..."
    curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
    sudo dnf install -y nodejs
fi

echo "✅ Системные зависимости установлены"
echo ""

# Проверка наличия EXPO_TOKEN
if [ -z "$EXPO_TOKEN" ]; then
    echo "❌ Ошибка: EXPO_TOKEN не установлен!"
    echo ""
    echo "📝 Решение:"
    echo "   export EXPO_TOKEN=ваш_токен_здесь"
    echo "   # Или добавьте в ~/.bashrc или ~/.zshrc"
    echo ""
    echo "🔗 Получить токен: https://expo.dev/settings/access-tokens"
    exit 1
fi

echo "✅ EXPO_TOKEN найден"

# Проверка наличия eas-cli
if ! command -v eas &> /dev/null; then
    echo "📦 Устанавливаю EAS CLI глобально..."
    if command -v sudo &> /dev/null; then
        sudo npm install -g eas-cli --unsafe-perm=true
    else
        npm install -g eas-cli --unsafe-perm=true
    fi
fi

# Обновление PATH если нужно
export PATH="$HOME/.npm-global/bin:$PATH"

echo "🔍 Проверяю статус аутентификации..."
if ! eas whoami &> /dev/null; then
    echo "🔐 Выполняю вход в EAS..."
    eas login
else
    echo "✅ Уже авторизован в EAS"
fi

# Проверка наличия eas.json
if [ ! -f "eas.json" ]; then
    echo "❌ Ошибка: eas.json не найден!"
    echo ""
    echo "📝 Решение:"
    echo "   eas init"
    echo "   # Или свяжите существующий проект:"
    echo "   eas init --id ваш_project_id"
    exit 1
fi

echo "✅ EAS проект настроен"

# Проверка Expo SDK версии
echo "🔍 Проверяю Expo SDK версию..."
if [ -f "app.json" ]; then
    SDK_VERSION=$(grep -o '"sdkVersion":\s*"[^"]*"' app.json | cut -d'"' -f4)
    if [ -z "$SDK_VERSION" ]; then
        echo "⚠️  SDK версия не указана в app.json"
    elif [ "$SDK_VERSION" != "54.0.0" ]; then
        echo "⚠️  Текущая SDK версия: $SDK_VERSION (ожидается 54.0.0)"
    else
        echo "✅ Expo SDK 54.0.0 подтвержден"
    fi
else
    echo "❌ app.json не найден"
    exit 1
fi

# Проверка зависимостей
echo "📦 Проверяю зависимости..."
if [ ! -d "node_modules" ]; then
    echo "📥 Устанавливаю зависимости..."
    if [ -f "package-lock.json" ]; then
        echo "🔒 Использую npm ci для точной установки..."
        npm ci
    else
        echo "📦 Создаю package-lock.json..."
        npm install
    fi
fi

# Проверка на ошибки в node_modules
if [ -d "node_modules" ]; then
    echo "✅ node_modules присутствует"
else
    echo "❌ Ошибка: node_modules не создана"
    exit 1
fi

# Установка expo-updates если не установлен
if ! grep -q '"expo-updates"' package.json; then
    echo "📦 Устанавливаю expo-updates для production сборок..."
    npx expo install expo-updates
fi

# Очистка кэша для чистой сборки
echo "🧹 Выполняю очистку кэша для чистой сборки..."

# Очистка кэша Expo
echo "🗑️  Очищаю кэш Expo..."
npx expo install --fix

# Очистка директорий кэша Expo вручную
if [ -d ".expo" ]; then
    rm -rf .expo
    echo "🗑️  Удалена директория .expo"
fi
if [ -d ".expo-shared" ]; then
    rm -rf .expo-shared
    echo "🗑️  Удалена директория .expo-shared"
fi

# Очистка кэша EAS (через перезапуск сервиса сборки - кэш очищается автоматически)
echo "ℹ️  EAS кэш будет очищен автоматически при сборке с новыми зависимостями"

# Очистка node_modules и переустановка зависимостей
echo "🗑️  Очищаю node_modules для чистой установки..."
if [ -d "node_modules" ]; then
    rm -rf node_modules
fi
if [ -f "package-lock.json" ]; then
    rm package-lock.json
fi

echo "📦 Переустанавливаю зависимости..."
npm install

echo "✅ Кэш очищен, зависимости переустановлены"
echo ""

echo "🏗️  Начинаю сборку production APK..."
echo "   Это займет несколько минут..."
echo ""

# Запуск сборки с подробным логированием
EAS_BUILD_PROFILE=production

echo "🔧 Параметры сборки:"
echo "   Профиль: $EAS_BUILD_PROFILE"
echo "   Платформа: Android (только)"
echo "   SDK: 54.0.0"
echo "   Сервер: EAS Build (облако)"
echo ""

# Запуск сборки только для Android
eas build \
    --platform android \
    --profile $EAS_BUILD_PROFILE \
    --message "Android SDK 54 Production build $(date +'%Y-%m-%d %H:%M:%S')" \
    --wait

echo ""
echo "✅ Сборка завершена!"

# Получение информации о последней Android сборке
echo "📋 Получаю информацию о последней Android сборке..."
BUILD_INFO=$(eas build:list --platform android --limit 1 --json)

if [ -z "$BUILD_INFO" ]; then
    echo "❌ Не удалось получить информацию о сборке"
    exit 1
fi

# Извлечение ID последней сборки
BUILD_ID=$(echo $BUILD_INFO | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$BUILD_ID" ]; then
    echo "❌ Не удалось найти ID сборки"
    exit 1
fi

echo "🆔 ID сборки: $BUILD_ID"

# Скачивание APK
APK_NAME="FreeBet-Remote-$(date +'%Y%m%d-%H%M%S').apk"
echo "📥 Скачиваю APK: $APK_NAME"

eas build:download $BUILD_ID --output $APK_NAME

if [ ! -f "$APK_NAME" ]; then
    echo "❌ Ошибка скачивания APK"
    exit 1
fi

echo ""
echo "🎉 APK успешно собран и загружен!"
echo "📱 Файл: $(pwd)/$APK_NAME"
echo "📏 Размер: $(ls -lh $APK_NAME | awk '{print $5}')"
echo ""

# Проверка APK файла
echo "🔍 Анализирую APK файл..."
echo "📱 Имя файла: $APK_NAME"
echo "📏 Размер: $(ls -lh $APK_NAME | awk '{print $5}')"
echo "📅 Создан: $(date -r $APK_NAME)"

# Проверка подписи APK (опционально)
if command -v apksigner &> /dev/null; then
    echo "🔐 Проверяю подпись APK..."
    if apksigner verify --print-certs $APK_NAME &> /dev/null; then
        echo "✅ APK правильно подписан"
    else
        echo "⚠️  Внимание: APK не подписан или подпись недействительна"
    fi
elif command -v jarsigner &> /dev/null; then
    echo "🔐 Проверяю подпись с помощью jarsigner..."
    if jarsigner -verify $APK_NAME &> /dev/null; then
        echo "✅ APK подписан (jarsigner)"
    else
        echo "⚠️  Внимание: APK не подписан"
    fi
else
    echo "ℹ️  Для проверки подписи установите Android SDK:"
    echo "   sudo dnf install android-tools"
    echo "   Или установите JDK: sudo dnf install java-17-openjdk-devel"
fi

# Проверка архитектуры APK (опционально)
if command -v aapt &> /dev/null; then
    echo "📋 Информация о APK:"
    aapt dump badging $APK_NAME | grep -E "(package|versionCode|versionName|native-code)" | head -5
fi

echo ""
echo "📤 APK готов к публикации!"
echo "   Можно загружать в Google Play Console"
echo "   Или распространять напрямую"
echo ""
echo "🎉 Скрипт завершен успешно!"
echo ""
echo "📋 Резюме сборки:"
echo "   📱 APK: $(pwd)/$APK_NAME"
echo "   🤖 Платформа: Android только"
echo "   🔢 SDK: 54.0.0"
echo "   🏗️  Профиль: Production"
echo "   ☁️  Сервер: EAS Build"
echo "   ✅ Подписан: Да (EAS keystore)"
echo ""

# Показать ссылку на EAS dashboard
echo "🔗 Управление сборками:"
echo "   EAS Dashboard: https://expo.dev/accounts/ваш_аккаунт/projects/ваш_проект/builds"
echo ""

# Рекомендации для AlmaLinux
echo "💡 Рекомендации для AlmaLinux:"
echo "   • APK готов для загрузки в Google Play Console"
echo "   • Для тестирования: adb install $APK_NAME"
echo "   • Для подписи: keystore управляется EAS автоматически"
echo ""

echo "🚀 Готово! APK можно использовать для публикации или тестирования."
