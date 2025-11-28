import { DynamoDBClient, CreateTableCommand, ListTablesCommand } from '@aws-sdk/client-dynamodb';
import dotenv from 'dotenv';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
dotenv.config({ path: join(__dirname, '..', '.env') });

const client = new DynamoDBClient({
  region: process.env.AWS_REGION,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
  },
});

const tables = [
  {
    TableName: process.env.DYNAMO_USERS_TABLE || 'Users',
    KeySchema: [
      { AttributeName: 'userId', KeyType: 'HASH' },
    ],
    AttributeDefinitions: [
      { AttributeName: 'userId', AttributeType: 'S' },
      { AttributeName: 'email', AttributeType: 'S' },
    ],
    GlobalSecondaryIndexes: [
      {
        IndexName: 'EmailIndex',
        KeySchema: [
          { AttributeName: 'email', KeyType: 'HASH' },
        ],
        Projection: {
          ProjectionType: 'ALL',
        },
        ProvisionedThroughput: {
          ReadCapacityUnits: 1,
          WriteCapacityUnits: 1,
        },
      },
    ],
    ProvisionedThroughput: {
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
  },
  {
    TableName: process.env.DYNAMO_ROOMS_TABLE || 'Rooms',
    KeySchema: [
      { AttributeName: 'roomId', KeyType: 'HASH' },
    ],
    AttributeDefinitions: [
      { AttributeName: 'roomId', AttributeType: 'S' },
    ],
    ProvisionedThroughput: {
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
  },
  {
    TableName: process.env.DYNAMO_MESSAGES_TABLE || 'Messages',
    KeySchema: [
      { AttributeName: 'roomId', KeyType: 'HASH' },
      { AttributeName: 'timestamp', KeyType: 'RANGE' },
    ],
    AttributeDefinitions: [
      { AttributeName: 'roomId', AttributeType: 'S' },
      { AttributeName: 'timestamp', AttributeType: 'S' },
    ],
    ProvisionedThroughput: {
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
  },
  {
    TableName: process.env.DYNAMO_CODESYNC_TABLE || 'CodeSync',
    KeySchema: [
      { AttributeName: 'roomId', KeyType: 'HASH' },
    ],
    AttributeDefinitions: [
      { AttributeName: 'roomId', AttributeType: 'S' },
    ],
    ProvisionedThroughput: {
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
  },
];

async function setupDynamoDB() {
  try {
    console.log('🔍 Checking existing tables...');
    const listCommand = new ListTablesCommand({});
    const existingTables = await client.send(listCommand);
    
    for (const table of tables) {
      if (existingTables.TableNames?.includes(table.TableName)) {
        console.log(`Table "${table.TableName}" already exists`);
      } else {
        console.log(`Creating table "${table.TableName}"...`);
        const createCommand = new CreateTableCommand(table);
        await client.send(createCommand);
        console.log(`Table "${table.TableName}" created successfully`);
      }
    }
    
    console.log('\nDynamoDB setup complete!');
    console.log('\nTables created:');
    tables.forEach(t => console.log(`  - ${t.TableName}`));
    
  } catch (error) {
    console.error('Error setting up DynamoDB:', error);
    process.exit(1);
  }
}

setupDynamoDB();
