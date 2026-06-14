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

After the AAR is available, wire `LibboxProxyEngine` to the generated Android API. The rest of the Android app already supplies:

- the selected Typhoon relay,
- the generated sing-box JSON config,
- the Android VPN tunnel file descriptor,
- the active `VpnService` for socket protection and lifecycle callbacks.
