#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run Allero."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

# show notification to change execution policy:
$allowedExecutionPolicy = @('Unrestricted', 'RemoteSigned', 'ByPass')
if ((Get-ExecutionPolicy).ToString() -notin $allowedExecutionPolicy) {
    Write-Output "PowerShell requires an execution policy in [$($allowedExecutionPolicy -join ", ")] to run Allero."
    Write-Output "For example, to set the execution policy to 'RemoteSigned' please run :"
    Write-Output "'Set-ExecutionPolicy RemoteSigned -scope CurrentUser'"
    break
}

$osArchitecture = if([Environment]::Is64BitProcess) { 'x86_64' } else { '386' }
$DOWNLOAD_URL = (Invoke-WebRequest -Uri 'https://api.github.com/repos/allero-io/allero/releases/latest' -UseBasicParsing | select-string -Pattern "https://github.com/allero-io/allero/releases/download/\d+\.\d+\.\d+/allero_\d+\.\d+\.\d+_windows_$osArchitecture.zip").Matches.Value
$OUTPUT_BASENAME = "allero-latest"
$OUTPUT_BASENAME_WITH_POSTFIX = "$OUTPUT_BASENAME.zip"

Write-Host 'Installing Allero...'
Write-Host ''
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $OUTPUT_BASENAME_WITH_POSTFIX -UseBasicParsing
Write-Host "[V] Downloaded Allero" -ForegroundColor DarkGreen

Expand-Archive -Path $OUTPUT_BASENAME_WITH_POSTFIX -DestinationPath $OUTPUT_BASENAME -Force | Out-Null

$localAppDataPath = $env:LOCALAPPDATA
$alleroPath = Join-Path "$localAppDataPath" 'allero'
New-Item -ItemType Directory -Force -Path $alleroPath | Out-Null

Copy-Item "$OUTPUT_BASENAME/*" -Destination "$alleroPath/" -Recurse -Force | Out-Null

Remove-Item -Recurse $OUTPUT_BASENAME

Write-Host "[V] Finished Installation" -ForegroundColor DarkGreen
Write-Host ""
Write-Host "To run allero globally, please follow these steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "    1. Run the following command as administrator: ``setx PATH `"`$env:path;$alleroPath`" -m``"
Write-Host ""
Write-Host "    2. Close and reopen your terminal."
Write-Host ""
