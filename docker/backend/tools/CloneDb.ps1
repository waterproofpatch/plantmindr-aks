Add-Type -Path "C:\path\to\Npgsql.dll"

$sourceConnection = New-Object Npgsql.NpgsqlConnection($sourceConnectionString)
$targetConnection = New-Object Npgsql.NpgsqlConnection($targetConnectionString)
$sourceConnection.Open()
$targetConnection.Open()

# Get list of tables from the source database
$command = $sourceConnection.CreateCommand()
$command.CommandText = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
$reader = $command.ExecuteReader()

$tables = @()
while ($reader.Read()) {
    $tables += $reader["table_name"]
}
$reader.Close()

# Copy each table
foreach ($table in $tables) {
    # Create table in the target database
    $createTableCommand = $sourceConnection.CreateCommand()
    $createTableCommand.CommandText = "SELECT 'CREATE TABLE ' || table_name || ' (' || string_agg(column_name || ' ' || data_type, ', ') || ');' FROM information_schema.columns WHERE table_name = '$table' GROUP BY table_name"
    $createTableSql = $createTableCommand.ExecuteScalar()
    
    $targetCreateCommand = $targetConnection.CreateCommand()
    $targetCreateCommand.CommandText = $createTableSql
    $targetCreateCommand.ExecuteNonQuery()

    # Copy data from source to target
    $copyCommand = $sourceConnection.CreateCommand()
    $copyCommand.CommandText = "COPY $table TO STDOUT WITH CSV"
    $copyReader = $copyCommand.ExecuteReader()

    $insertCommand = $targetConnection.CreateCommand()
    $insertCommand.CommandText = "COPY $table FROM STDIN WITH CSV"
    $insertWriter = $insertCommand.ExecuteNonQuery()

    while ($copyReader.Read()) {
        $insertWriter.WriteLine($copyReader.GetValue(0))
    }

    $copyReader.Close()
    $insertWriter.Close()
}

$sourceConnection.Close()
$targetConnection.Close()
