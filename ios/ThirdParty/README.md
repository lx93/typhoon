# Third-party iOS engine artifacts

`Libbox.xcframework` is generated locally from sing-box and intentionally ignored by git because it is large.

To rebuild it:

```sh
git clone --depth 1 --branch testing https://github.com/SagerNet/sing-box.git /private/tmp/typhoon-sing-box
cd /private/tmp/typhoon-sing-box
go install github.com/sagernet/gomobile/cmd/gomobile@v0.1.12
go install github.com/sagernet/gomobile/cmd/gobind@v0.1.12
PATH="$HOME/go/bin:$PATH" gomobile init
PATH="$HOME/go/bin:$PATH" go run ./cmd/internal/build_libbox -target apple -platform ios,iossimulator -debug
mkdir -p /Users/lx93/Documents/typhoon/ios/ThirdParty
cp -R Libbox.xcframework /Users/lx93/Documents/typhoon/ios/ThirdParty/Libbox.xcframework
```
