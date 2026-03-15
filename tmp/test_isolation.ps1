
try {
    Write-Host "--- Registering User A ---"
    $resA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method Post -Body '{"username":"user_a_test","password":"password"}' -ContentType "application/json"
} catch {
    Write-Host "User A registration failed or already exists. Proceeding to login."
}

Write-Host "--- Logging in User A ---"
$loginA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body '{"username":"user_a_test","password":"password"}' -ContentType "application/json"
$tokenA = $loginA.token
Write-Host "User A Token: $tokenA"

try {
    Write-Host "--- Registering User B ---"
    $resB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method Post -Body '{"username":"user_b_test","password":"password"}' -ContentType "application/json"
} catch {
    Write-Host "User B registration failed or already exists. Proceeding to login."
}

Write-Host "--- Logging in User B ---"
$loginB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body '{"username":"user_b_test","password":"password"}' -ContentType "application/json"
$tokenB = $loginB.token
Write-Host "User B Token: $tokenB"

Write-Host "--- User A creating Task A ---"
$taskA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Post -Body '{"title":"TASK_FOR_USER_A","dimension":"work","template":"task"}' -ContentType "application/json" -Headers @{Authorization="Bearer $tokenA"}
$taskID_A = $taskA.id
Write-Host "Created Task A ID: $taskID_A"

Write-Host "--- User B listing tasks (Should be empty) ---"
$tasksB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Get -Headers @{Authorization="Bearer $tokenB"}
Write-Host "User B tasks count: $($tasksB.Count)"
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
    Write-Host "PASS: User B blocked from Task A. Error: $($_.Exception.Message)"
}

Write-Host "--- User B creating Task B ---"
$taskB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Post -Body '{"title":"TASK_FOR_USER_B","dimension":"work","template":"task"}' -ContentType "application/json" -Headers @{Authorization="Bearer $tokenB"}
Write-Host "Created Task B ID: $($taskB.id)"

Write-Host "--- User A listing tasks (Should only see 1 task: Task A) ---"
$tasksA_Final = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" -Method Get -Headers @{Authorization="Bearer $tokenA"}
Write-Host "User A tasks count: $($tasksA_Final.Count)"
$foundB = $false
foreach ($t in $tasksA_Final) {
    if ($t.title -eq "TASK_FOR_USER_B") { $foundB = $true }
}
if ($foundB) {
    Write-Host "FAIL: User A sees User B's task!"
} else {
    Write-Host "PASS: User A data is isolated."
}
