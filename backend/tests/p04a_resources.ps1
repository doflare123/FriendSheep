param(
    [Parameter(Mandatory=$true)][int]$PostmasterId,
    [Parameter(Mandatory=$true)][string]$OutputPath,
    [Parameter(Mandatory=$true)][string]$StopPath,
    [int]$MaxMinutes=120
)

$ErrorActionPreference='Stop'
$postmaster=Get-CimInstance Win32_Process -Filter "ProcessId=$PostmasterId"
if (-not $postmaster -or -not $postmaster.CommandLine.Contains('D:/friendSheep/backend/.cache/p04a-20261025-postgres')) {
    throw 'Процесс не принадлежит изолированному кластеру P0.4a'
}
$cpu=Get-CimInstance Win32_Processor | Select-Object -First 1
$computer=Get-CimInstance Win32_ComputerSystem
[pscustomobject]@{kind='environment';timestamp=(Get-Date).ToUniversalTime().ToString('o');cpu_name=$cpu.Name;logical_cpus=$computer.NumberOfLogicalProcessors;physical_memory_bytes=$computer.TotalPhysicalMemory;postmaster_pid=$PostmasterId;scope='Только изолированный postmaster и его дочерние PostgreSQL процессы. WorkingSet суммируется с возможным повторным учётом общей памяти; это не PSS.'} | ConvertTo-Json -Compress | Add-Content -LiteralPath $OutputPath -Encoding UTF8
$deadline=(Get-Date).AddMinutes($MaxMinutes)
while ((Get-Date) -lt $deadline -and -not (Test-Path -LiteralPath $StopPath)) {
    $members=Get-CimInstance Win32_Process -Filter "Name='postgres.exe'" | Where-Object { $_.ProcessId -eq $PostmasterId -or $_.ParentProcessId -eq $PostmasterId }
    $samples=@(foreach($member in $members) {
        $process=Get-Process -Id $member.ProcessId -ErrorAction SilentlyContinue
        if($process) { [pscustomobject]@{pid=$member.ProcessId;cpu_seconds=$process.CPU;working_set_bytes=$process.WorkingSet64;private_bytes=$process.PrivateMemorySize64} }
    })
    [pscustomobject]@{kind='sample';timestamp=(Get-Date).ToUniversalTime().ToString('o');processes=$samples} | ConvertTo-Json -Depth 4 -Compress | Add-Content -LiteralPath $OutputPath -Encoding UTF8
    Start-Sleep -Seconds 2
}
