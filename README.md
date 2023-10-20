# ngrok4com

hub4com GUI and KiTTY helper for configuring remote devices via ngrok by serial port

## Credits

- Vyacheslav Frolov - for [com0com and hub4com](https://com0com.sourceforge.net)
- Simon Tatham, Cyril Dupont - for [KiTTY](https://github.com/cyd01/KiTTY)
- ngrok - for [ngrok](https://github.com/ngrok/ngrok-go)

## Usage:

- git clone https://github.com/abakum/ngrok4com
- place NGROK_AUTHTOKEN.txt to ngrok4com\ before build for embed or set env during run
- place NGROK_API_KEY.txt to ngrok4com\ before build for embed or set env during run
- place hub4com.exe to ngrok4com\bin\ before build for embed
- place kitty_portable.exe to ngrok4com\bin\ before build for embed
- on remote PC with COM7
  - run `ngrok4com 7`
- on local PC
  - install [com0com](https://sourceforge.net/projects/com0com/files/com0com/3.0.0.0)
  - setup COM11 for KiTTY by running<br>
 `com0com\setupc.exe install 0 PortName=COM#,RealPortName=COM11,EmuBR=yes,AddRTTO=1,AddRITO=1 -`
  - run `ngrok4com` for control remote device over COM7 of remote PC

[Look more](RFC2217.md)
