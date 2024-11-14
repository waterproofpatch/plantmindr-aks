using System;
using Npgsql;

class Program
{
	static void Main()
	{
		string? sourceConnectionString = Environment.GetEnvironmentVariable("DB_CONN_STRING");
		string? destinationConnectionString = Environment.GetEnvironmentVariable("DESTINATION_CONNECTION_STRING");
		if (sourceConnectionString == null || destinationConnectionString == null)
		{
			Console.WriteLine("Not all environment variables are initialized!");
			return;
		}
		Console.WriteLine("Got environment variables...");

		using (var sourceConnection = new NpgsqlConnection(sourceConnectionString))
		using (var destinationConnection = new NpgsqlConnection(destinationConnectionString))
		{
			sourceConnection.Open();
			destinationConnection.Open();

			Console.WriteLine("Opened source and destination connections...");

			// Get the list of tables in the source database
			using (var sourceCommand = new NpgsqlCommand("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'", sourceConnection))
			using (var reader = sourceCommand.ExecuteReader())
			{
				Console.WriteLine("Tables in the source database:");
				while (reader.Read())
				{
					string tableName = reader["table_name"].ToString();
					Console.WriteLine("Read table name: " + tableName);

					// Create the table in the destination database if it doesn't exist
					using (var createTableCommand = new NpgsqlCommand($@"
                        CREATE TABLE IF NOT EXISTS {tableName} AS TABLE {tableName} WITH NO DATA", destinationConnection))
					{
						createTableCommand.ExecuteNonQuery();
					}

					// Copy data from the source table to the destination table
					using (var sourceDataCommand = new NpgsqlCommand($"SELECT * FROM {tableName}", sourceConnection))
					using (var dataReader = sourceDataCommand.ExecuteReader())
					{
						while (dataReader.Read())
						{
							var columnNames = new string[dataReader.FieldCount];
							var columnValues = new object[dataReader.FieldCount];

							for (int i = 0; i < dataReader.FieldCount; i++)
							{
								columnNames[i] = dataReader.GetName(i);
								columnValues[i] = dataReader.GetValue(i);
							}

							var insertCommandText = $"INSERT INTO {tableName} ({string.Join(", ", columnNames)}) VALUES ({string.Join(", ", columnNames.Select((_, i) => $"@p{i}"))})";
							using (var destinationCommand = new NpgsqlCommand(insertCommandText, destinationConnection))
							{
								for (int i = 0; i < columnNames.Length; i++)
								{
									destinationCommand.Parameters.AddWithValue($"@p{i}", columnValues[i]);
								}
								destinationCommand.ExecuteNonQuery();
							}
						}
					}
				}
			}
		}
	}
}
