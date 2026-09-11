IC-SDR Go - portable distribution
=================================

*English translation of DISTRIBUTION.md. The Spanish original remains the
reference document.*

Copy the whole IC-SDR-Go folder and run IC-SDR-Go.exe, keeping DATA alongside
the executable. The package includes the Visual C++ runtime that SoapySDR and
RTL-SDR need, so there is no need to install it separately.
DATA holds the SoapySDR/SDRplay, DMR, RTL_433 and APRS runtimes together with
their supporting data and licences. The program does not have to be started
from any particular folder.

RADIOSONDES covers RS41, DFM and M10/M20 from rs1729/RS. Select a family, tune
and press START. Their sources, GPL-3.0 licence and build instructions sit
beside the executables in DATA\tools\radiosonde\runtime.
CSV and JSON files are written to DATA\exports\radiosonde.

AIS VESSELS integrates AIS-catcher and covers both the 161.975 and 162.025 MHz
channels at once. The OPEN MAP button shows, in a separate window, the
positions, course, speed and identification data received directly over the
air.

ADS-B AIRCRAFT lets you select 1090 MHz (ADS-B/Mode S) or 978 MHz (UAT), with
an aircraft list and a separate map showing positions and trails.

Settings, memories, recordings, captures, exports and logs are all kept inside
DATA. If the program never reaches its interface, check
DATA\logs\startup.log; unrecoverable failures also raise a message box.

Language: IC-SDR starts in Spanish or English according to the Windows locale,
and the flags in the bottom settings strip switch between them at any time. The
choice is stored in DATA\config\settings.json.

To rebuild this distribution from source:
    powershell -ExecutionPolicy Bypass -File .\build-release.ps1
