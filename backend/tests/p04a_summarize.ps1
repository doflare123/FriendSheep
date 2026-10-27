param([Parameter(Mandatory=$true)][string]$Directory)

$ErrorActionPreference='Stop'
$previousCpu=@{}
$samples=[System.Collections.Generic.List[object]]::new()
$environment=$null
function Convert-P04aTimestamp($value) {
    if($value -is [DateTime]) { return [DateTimeOffset]$value }
    return [DateTimeOffset]::Parse([string]$value,[Globalization.CultureInfo]::InvariantCulture)
}
foreach($line in Get-Content -LiteralPath (Join-Path $Directory 'resource_samples.jsonl') -Encoding UTF8) {
    $record=$line | ConvertFrom-Json
    if($record.kind -eq 'environment') { $environment=$record; continue }
    $cpuDelta=0.0
    $workingSet=0L
    $privateBytes=0L
    foreach($process in $record.processes) {
        $key=[string]$process.pid
        $cpu=[double]$process.cpu_seconds
        if($previousCpu.ContainsKey($key)) {
            $delta=$cpu-$previousCpu[$key]
            # Уменьшение CPU означает повторное использование PID новым процессом.
            if($delta -ge 0) { $cpuDelta += $delta } else { $cpuDelta += $cpu }
        }
        elseif($samples.Count -gt 0) { $cpuDelta += $cpu }
        $previousCpu[$key]=$cpu
        $workingSet += [long]$process.working_set_bytes
        $privateBytes += [long]$process.private_bytes
    }
    $samples.Add([pscustomobject]@{time=(Convert-P04aTimestamp $record.timestamp);cpu_delta_s=$cpuDelta;working_set_bytes=$workingSet;private_bytes=$privateBytes})
}

$results=@(foreach($file in Get-ChildItem -LiteralPath $Directory -Filter 'n*.json' | Sort-Object Name) {
    $case=Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8 | ConvertFrom-Json
    $start=Convert-P04aTimestamp $case.workload_started_at
    $end=Convert-P04aTimestamp $case.workload_ended_at
    $selected=@($samples | Where-Object { $_.time -ge $start -and $_.time -le $end })
    $cpuSeconds=($selected | Measure-Object cpu_delta_s -Sum).Sum
    $duration=($end-$start).TotalSeconds
    $single=($case.single_plan | Select-String '^Execution Time:').Line -replace '^Execution Time: ([0-9.]+) ms$','$1'
    $batch=($case.batch8_plan | Select-String '^Execution Time:').Line -replace '^Execution Time: ([0-9.]+) ms$','$1'
    [pscustomobject]@{
        case=$file.BaseName
        participants=$case.participants_per_event
        history_per_cohort=$case.history_per_cohort
        max_history_per_user=$case.max_history_per_user
        users=$case.users
        process_s=$case.process_6_events_ms/1000.0
        saturation_drain_s=$case.drain_ms/1000.0
        calculations_per_second=$case.calculations_per_second
        default_polling_estimate_s=$case.default_polling_total_ms_estimate/1000.0
        baseline_profile_p95_ms=$case.baseline_http.profile.p95_ms
        baseline_profile_p99_ms=$case.baseline_http.profile.p99_ms
        baseline_event_p95_ms=$case.baseline_http.event.p95_ms
        baseline_event_p99_ms=$case.baseline_http.event.p99_ms
        profile_p95_ms=$case.loaded_http.profile.p95_ms
        profile_p99_ms=$case.loaded_http.profile.p99_ms
        event_p95_ms=$case.loaded_http.event.p95_ms
        event_p99_ms=$case.loaded_http.event.p99_ms
        http_errors=$case.loaded_http.errors
        pool_wait_count=$case.pool.WaitCount
        temp_files_delta=[long]$case.db_after.temp_files-[long]$case.db_before.temp_files
        temp_bytes_delta=[long]$case.db_after.temp_bytes-[long]$case.db_before.temp_bytes
        membership_bytes=$case.db_after.membership_bytes
        queue_bytes=$case.db_after.queue_bytes
        queue_growth_bytes=[long]$case.db_after.queue_bytes-[long]$case.db_before.queue_bytes
        aggregate_growth_bytes=[long]$case.db_after.aggregate_bytes-[long]$case.db_before.aggregate_bytes
        single_sql_ms=[double]::Parse($single,[Globalization.CultureInfo]::InvariantCulture)
        batch8_sql_ms=[double]::Parse($batch,[Globalization.CultureInfo]::InvariantCulture)
        resource_samples=$selected.Count
        postgres_cpu_seconds=$cpuSeconds
        postgres_average_logical_cores=if($duration -gt 0){$cpuSeconds/$duration}else{0}
        postgres_peak_working_set_mib=($selected | Measure-Object working_set_bytes -Maximum).Maximum/1MB
        postgres_peak_private_mib=($selected | Measure-Object private_bytes -Maximum).Maximum/1MB
    }
})
[pscustomobject]@{environment=$environment;cases=$results;resource_limitations='Сэмплы раз в 2s; CPU interval обрезается по timestamp с погрешностью границ. Короткоживущие процессы между сэмплами могут не учитываться. WorkingSet включает общую память повторно; private bytes — committed private memory, не точный resident RAM. Метрики относятся к изолированному PostgreSQL кластеру, не Go/HTTP.'} |
    ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $Directory 'summary.json') -Encoding UTF8
$results | Select-Object case,users,process_s,saturation_drain_s,profile_p99_ms,event_p99_ms,postgres_average_logical_cores,postgres_peak_working_set_mib | Format-Table -AutoSize
