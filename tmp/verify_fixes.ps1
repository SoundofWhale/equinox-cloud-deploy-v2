
$timestamp = Get-Date -Format "HHmmss"
$userE = "user_e_$timestamp"

Write-Host "--- Registering and Logging in User E ($userE) ---"
$resE = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method Post -Body (@{username=$userE; password="password"} | ConvertTo-Json) -ContentType "application/json"
$loginE = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body (@{username=$userE; password="password"} | ConvertTo-Json) -ContentType "application/json"
$tokenE = $loginE.token

Write-Host "--- Attempting to Create Task (Testing Authorization Header fix) ---"
try {
    # This simulates the frontend call that was failing
    $taskE = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Post -Body (@{title="VERIFY_FIX_TASK"; dimension="work"; template="task"} | ConvertTo-Json) -ContentType "application/json" -Headers @{Authorization="Bearer $tokenE"}
    Write-Host "SUCCESS: Task created with ID: $($taskE.id)"
} catch {
    Write-Host "FAILURE: Could not create task. Status: $($_.Exception.Response.StatusCode)"
    return
}

Write-Host "--- Verifying GetContextPacket (Testing SQL Scan fix) ---"
try {
    # dimension=work is the default the dashboard uses
    $context = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/context?dimension=work" -Method Get -Headers @{Authorization="Bearer $tokenE"}
    $found = $false
    foreach ($t in $context.data.tasks) {
        if ($t.title -eq "VERIFY_FIX_TASK") { $found = $true }
    }
    if ($found) {
        Write-Host "PASS: Task found in context packet."
    } else {
        Write-Host "FAIL: Task NOT found in context packet. Tasks returned: $($context.data.tasks.Count)"
    }
} catch {
    Write-Host "FAILURE: GetContextPacket failed. Error: $($_.Exception.Message)"
}
