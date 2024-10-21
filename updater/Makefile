.PHONY: all windows-x86 macos-x86 macos-arm clean

# 기본 all 타겟
all: windows-x86 macos-x86 macos-arm

# Windows x86 빌드
windows-x86:
	GOOS=windows GOARCH=amd64 go build -o bambu-updater-windows-x86.exe main.go

# macOS x86 빌드
macos-x86:
	GOOS=darwin GOARCH=amd64 go build -o bambu-updater-macos-x86 main.go

# macOS ARM (Apple Silicon) 빌드
macos-arm:
	GOOS=darwin GOARCH=arm64 go build -o bambu-updater-macos-arm main.go

# 빌드된 파일 정리
clean:
	rm -f bambu-updater-windows-x86.exe bambu-updater-macos-x86 bambu-updater-macos-arm
