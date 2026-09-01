[CmdletBinding()]
param(
    [string] $BinaryPath = "bin\logbasset.exe",
    [string] $ExpectedVersion
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $BinaryPath -PathType Leaf)) {
    throw "smoke-test: FAIL - binary not found or not a file: $BinaryPath"
}

$ResolvedBinaryPath = (Resolve-Path -LiteralPath $BinaryPath -ErrorAction Stop).Path

function Invoke-SmokeCommand {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string] $CheckName,
        [Parameter(Mandatory = $true)]
        [string[]] $Arguments
    )

    try {
        $Output = @(& $ResolvedBinaryPath @Arguments 2>&1)
    }
    catch {
        throw "smoke-test: FAIL - $CheckName could not execute: $($_.Exception.Message)"
    }

    $ExitCode = $LASTEXITCODE
    if ($ExitCode -ne 0) {
        $Details = ($Output | ForEach-Object { $_.ToString() }) -join [Environment]::NewLine
        if ([string]::IsNullOrWhiteSpace($Details)) {
            $Details = "no output"
        }
        throw "smoke-test: FAIL - $CheckName exited with code ${ExitCode}: $Details"
    }

    Write-Host "smoke-test: ok - $CheckName"
    return $Output
}

$CoreCommands = @(
    "query",
    "power-query",
    "numeric-query",
    "facet-query",
    "timeseries-query",
    "tail"
)

$VersionOutput = @(Invoke-SmokeCommand -CheckName "--version" -Arguments @("--version"))
$VersionLines = @($VersionOutput | ForEach-Object { $_.ToString() })
$VersionText = $VersionLines -join [Environment]::NewLine

if ($PSBoundParameters.ContainsKey("ExpectedVersion")) {
    $ExpectedVersionText = "logbasset version $ExpectedVersion"
    if ($VersionLines.Count -ne 1 -or $VersionLines[0] -cne $ExpectedVersionText) {
        throw "smoke-test: FAIL - --version output unexpected: $VersionText"
    }
}
elseif (-not $VersionText.StartsWith("logbasset version ", [StringComparison]::Ordinal)) {
    throw "smoke-test: FAIL - --version output unexpected: $VersionText"
}

Write-Host "smoke-test: ok - --version output"

$HelpOutput = @(Invoke-SmokeCommand -CheckName "--help" -Arguments @("--help"))
$HelpText = ($HelpOutput | ForEach-Object { $_.ToString() }) -join [Environment]::NewLine
foreach ($Command in $CoreCommands) {
    if (-not $HelpText.Contains($Command)) {
        throw "smoke-test: FAIL - --help missing command: $Command"
    }
}
Write-Host "smoke-test: ok - --help lists core commands"

foreach ($Command in $CoreCommands) {
    [void](Invoke-SmokeCommand -CheckName "$Command --help" -Arguments @($Command, "--help"))
}

[void](Invoke-SmokeCommand -CheckName "context" -Arguments @("context"))
[void](Invoke-SmokeCommand -CheckName "schema" -Arguments @("schema"))
[void](Invoke-SmokeCommand -CheckName "completion powershell" -Arguments @("completion", "powershell"))

Write-Host "smoke-test: all checks passed for $BinaryPath"
