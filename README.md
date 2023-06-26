# ngrok4com

hub4com GUI and KiTTY helper for configuring remote devices via ngrok by serial port

## Credits

- Vyacheslav Frolov - for [com0com and hub4com](com0com.sourceforge.net)
- Cyril Dupont - for [KiTTY](https://github.com/cyd01/KiTTY)
- ngrok - for [ngrok](https://github.com/ngrok/ngrok-go)

## Usage:

- git clone https://github.com/abakum/ngrok4com
- place NGROK_AUTHTOKEN.txt to ngrok4com\ before build or set env during run
- place NGROK_API_KEY.txt to ngrok4com\ before build or set env during run
- on remote PC with COM3
  - place hub4com\hub4com.exe to ..\ngrok4com.exe 
  - run `ngrok4com 3`
- on local PC with KiTTY
  - install [com0com](https://sourceforge.net/projects/com0com/files/com0com/3.0.0.0)
  - setup COM11 for KiTTY by `run com0com\setupc.exe install PortName=COM11,EmuBR=yes -`
  - place hub4com\hub4com.exe to ..\ngrok4com.exe
  - place kitty\kitty_portable.exe to ..\ngrok4com.exe
  - run `ngrok4com 9600` for control remote device over COM3 of remote PC
