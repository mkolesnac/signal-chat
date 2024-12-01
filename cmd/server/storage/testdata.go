package storage

var testResources = []Resource{
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "acc#123",
			SortKey:      "acc#123",
		},
		CreatedAt: GetTimestamp(),
	},
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "acc#123",
			SortKey:      "conv#abs",
		},
		CreatedAt: GetTimestamp(),
	},
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "conv#abs",
			SortKey:      "conv#abs",
		},
		CreatedAt: GetTimestamp(),
	},
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "conv#abs",
			SortKey:      "acc#123",
		},
		CreatedAt: GetTimestamp(),
	},
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "conv#abs",
			SortKey:      "msg#1",
		},
		CreatedAt: GetTimestamp(),
	},
	{
		PrimaryKey: PrimaryKey{
			PartitionKey: "conv#abs",
			SortKey:      "msg#2",
		},
		CreatedAt: GetTimestamp(),
	},
}
