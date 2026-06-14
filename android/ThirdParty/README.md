# Third-party Android engine artifacts

The Android client expects a local generated sing-box/libbox AAR at:

```text
android/app/libs/libbox.aar
```

The AAR is intentionally ignored by Git because it is generated and large.

The iOS client already uses the same local-artifact pattern for `ios/ThirdParty/Libbox.xcframework`.

## Build direction

Use sing-box's Android libbox build flow, then copy the generated AAR into:

```sh
mkdir -p /Users/lx93/Documents/typhoon/android/app/libs
cp /path/to/generated/libbox.aar /Users/lx93/Documents/typhoon/android/app/libs/libbox.aar
```

The local emulator build was validated with:

```sh
brew install openjdk@17 android-commandlinetools
JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home \
  sdkmanager --sdk_root=/Users/lx93/Library/Android/sdk 'ndk;29.0.14206865'

go install github.com/sagernet/gomobile/cmd/gomobile@v0.1.12
go install github.com/sagernet/gomobile/cmd/gobind@v0.1.12

ANDROID_HOME=/Users/lx93/Library/Android/sdk \
ANDROID_NDK_HOME=/Users/lx93/Library/Android/sdk/ndk/29.0.14206865 \
JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home \
PATH="$HOME/go/bin:/opt/homebrew/opt/openjdk@17/bin:$PATH" \
  gomobile init

git clone --depth 1 --branch testing https://github.com/SagerNet/sing-box.git /private/tmp/typhoon-sing-box-android
cd /private/tmp/typhoon-sing-box-android
ANDROID_HOME=/Users/lx93/Library/Android/sdk \
ANDROID_NDK_HOME=/Users/lx93/Library/Android/sdk/ndk/29.0.14206865 \
JAVA_HOME=/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home \
PATH="$HOME/go/bin:/opt/homebrew/opt/openjdk@17/bin:$PATH" \
  go run ./cmd/internal/build_libbox -target android -platform android/arm64 -debug

cp /private/tmp/typhoon-sing-box-android/libbox.aar /Users/lx93/Documents/typhoon/android/app/libs/libbox.aar
```

`ndk;27.3.13750724` failed locally while linking sing-box's prebuilt Cronet library with `unknown relocation (315)`. `ndk;29.0.14206865` built the arm64 emulator AAR successfully.

After the AAR is available, `LibboxProxyEngine` uses the generated Android API with:

- the selected Typhoon relay,
- the generated sing-box JSON config,
- the active `VpnService` for socket protection and lifecycle callbacks,
- a minimal `PlatformInterface` implementation that opens the Android VPN TUN when libbox starts.
