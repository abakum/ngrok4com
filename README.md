# ngrok4com![ngrok4com](winres\icon32.png )

ngrok4com is hub4com GUI and KiTTY helper for configuring remote devices by serial port<br>
Helps KiTTY communicate with a hub4com over firewalls via ngrok.com<br>
Помогает KiTTY взаимодействовать с hub4com через брандмауэры с помощью ngrok.com

## Credits

- Vyacheslav Frolov - for [com0com and hub4com](https://com0com.sourceforge.net) ![com](winres\com.png)
- Simon Tatham, Cyril Dupont - for [KiTTY](https://github.com/cyd01/KiTTY) ![KiTTY](winres\KiTTY.png)
- ngrok - for [ngrok](https://github.com/ngrok/ngrok-go) ![ngrok](winres\ngrok.png)
- Balázs Szalkai - for [Greenfish Icon Editor Pro](http://greenfishsoftware.org/gfie.php) ![gfie](winres\gfie.png )
- Katayama Hirofumi MZ - for [RisohEditor](https://github.com/katahiromz/RisohEditor) ![re](winres\re.png )

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
