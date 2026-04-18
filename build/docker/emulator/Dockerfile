FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

# Android SDK paths
ENV ANDROID_HOME=/opt/android
ENV ANDROID_SDK_ROOT=/opt/android
ENV PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulator

# Установка зависимостей
RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    openjdk-17-jdk \
    libglu1-mesa \
    libpulse0 \
    libnss3 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxi6 \
    libxtst6 \
    libxdamage1 \
    libxrandr2 \
    libgbm1 \
    libc6 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/android

# Скачивание и правильная установка cmdline-tools
RUN wget https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip \
 && unzip commandlinetools-linux-11076708_latest.zip \
 && rm commandlinetools-linux-11076708_latest.zip \
 && mv cmdline-tools cmdline-tools-temp \
 && mkdir -p cmdline-tools \
 && mv cmdline-tools-temp cmdline-tools/latest

# Проверка sdkmanager (очень полезно при билде)
RUN sdkmanager --list

# Принятие лицензий
RUN yes | sdkmanager --licenses

RUN sdkmanager --update

# Установка SDK компонентов
RUN sdkmanager --channel=3 \
    "platform-tools" \
    "platforms;android-30" \
    "system-images;android-30;google_apis;arm64-v8a" \
    "emulator"

# Создание AVD
RUN echo "no" | avdmanager create avd \
    -n test-avd \
    -k "system-images;android-30;google_apis;arm64-v8a" \
    --device "pixel"

# Порт ADB
EXPOSE 5555

# Запуск эмулятора
CMD emulator -avd test-avd \
    -no-window \
    -no-audio \
    -no-boot-anim \
    -gpu off \
    -accel off \
    -memory 1024 \
    -skip-adb-auth \
    -no-snapshot \
    -no-snapshot-save \
    -no-snapshot-load