
$timestamp = Get-Date -Format "HHmmss"
$userA = "user_a_$timestamp"
$userB = "user_b_$timestamp"

try {
    Write-Host "--- Registering User A ($userA) ---"
    $resA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method Post -Body (@{username=$userA; password="password"} | ConvertTo-Json) -ContentType "application/json"
} catch {
    Write-Host "User A registration failed. Error: $($_.Exception.Message)"
}

Write-Host "--- Logging in User A ---"
$loginA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body (@{username=$userA; password="password"} | ConvertTo-Json) -ContentType "application/json"
$tokenA = $loginA.token
Write-Host "User A Token: $tokenA"

try {
    Write-Host "--- Registering User B ($userB) ---"
    $resB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method Post -Body (@{username=$userB; password="password"} | ConvertTo-Json) -ContentType "application/json"
} catch {
    Write-Host "User B registration failed. Error: $($_.Exception.Message)"
}

Write-Host "--- Logging in User B ---"
$loginB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body (@{username=$userB; password="password"} | ConvertTo-Json) -ContentType "application/json"
$tokenB = $loginB.token
Write-Host "User B Token: $tokenB"

Write-Host "--- User A creating Task A ---"
$taskA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Post -Body (@{title="TASK_FOR_USER_A"; dimension="work"; template="task"} | ConvertTo-Json) -ContentType "application/json" -Headers @{Authorization="Bearer $tokenA"}
$taskID_A = $taskA.id
Write-Host "Created Task A ID: $taskID_A"

Write-Host "--- User B listing tasks (Should be empty array []) ---"
$tasksB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Get -Headers @{Authorization="Bearer $tokenB"}
Write-Host "User B tasks count: $($tasksB.Count)"
# In PowerShell, an empty array [] from JSON becomes a $null or has 0 count depending on parsing.
# We expect count 0.
if ($tasksB.Count -gt 0) {
    Write-Host "FAIL: User B sees User A's tasks!"
    foreach ($t in $tasksB) {
        Write-Host "  Found Task: $($t.title) ($($t.id))"
    }
} else {
    Write-Host "PASS: User B sees 0 tasks."
}

Write-Host "--- User B trying to access Task A by ID (Should fail with 404 or 403) ---"
try {
    $taskA_By_B = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks/$taskID_A" -Method Get -Headers @{Authorization="Bearer $tokenB"}
    Write-Host "FAIL: User B successfully accessed User A's task! Response: $($taskA_By_B.title)"
} catch {
    Write-Host "PASS: User B blocked from Task A. Status: $($_.Exception.Response.StatusCode)"
}

Write-Host "--- User B creating Task B ---"
$taskBResult = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Post -Body (@{title="TASK_FOR_USER_B"; dimension="work"; template="task"} | ConvertTo-Json) -ContentType "application/json" -Headers @{Authorization="Bearer $tokenB"}
Write-Host "Created Task B ID: $($taskBResult.id)"

Write-Host "--- User A listing tasks (Should only see 1 task: Task A) ---"
$tasksA_Final = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Get -Headers @{Authorization="Bearer $tokenA"}
Write-Host "User A tasks count: $($tasksA_Final.Count)"
$foundB = $false
# Note: if Count is 1, $tasksA_Final might not be an array but a single object in some PS versions.
# Forcing it to array for safety.
$tasksAArray = @($tasksA_Final)
foreach ($t in $tasksAArray) {
    if ($t.title -eq "TASK_FOR_USER_B") { $foundB = $true }
}
if ($foundB) {
    Write-Host "FAIL: User A sees User B's task!"
} else {
    Write-Host "PASS: User A data is isolated."
}
