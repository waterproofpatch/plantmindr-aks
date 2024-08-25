Write-Host "Installing Npgsql..."
# Install-Package Npgsql
Write-Host "Adding type to lib..."
# Add-Type -Path "C:\path\to\Npgsql.dll"
Write-Host "Reading connection string"
# $connectionString = "Host=your_host;Username=your_username;Password=your_password;Database=your_database"
$connectionString = (Get-Item Env:DB_CONN_STRING).Value
$connection = New-Object Npgsql.NpgsqlConnection($connectionString)
$connection.Open()
$command = $connection.CreateCommand()
# $command.CommandText = "SELECT * FROM items"
$command.CommandText = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
$reader = $command.ExecuteReader()

Write-Host "Reading..."
while ($reader.Read()) {
    Write-Output $reader["your_column"]
}

$reader.Close()
$connection.Close()


