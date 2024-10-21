from cx_Freeze import setup, Executable
import sys

build_exe_options = {
    "packages": ["flask", "flask_cors", "ssl", "paho.mqtt.client", "datetime", "threading", "queue", "signal", "sys"],
    "includes": [],
    "include_files": [],
    "excludes": []
}

base = None
if sys.platform == "win32":
    base = "Win32GUI"  # Windows GUI application

setup(
    name="bambu-proxy",
    version="0.1",
    description="bambu-proxy",
    options={"build_exe": build_exe_options},
    executables=[Executable("proxy.py", base=base)]
)
