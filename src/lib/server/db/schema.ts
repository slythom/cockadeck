import { integer, sqliteTable, text, primaryKey } from 'drizzle-orm/sqlite-core';

export const user = sqliteTable('user', {
	id: text('id').primaryKey(),
	age: integer('age'),
	username: text('username').notNull().unique(),
	passwordHash: text('password_hash').notNull()
});

export const session = sqliteTable('session', {
	id: text('id').primaryKey(),
	userId: text('user_id')
		.notNull()
		.references(() => user.id),
	expiresAt: integer('expires_at', { mode: 'timestamp' }).notNull()
});

export const collections = sqliteTable('collections', {
	id: text('id').primaryKey(),
	userId: text('user_id')
		.notNull()
		.references(() => user.id),
	name: text('name').notNull()
});

export const cards = sqliteTable('cards', {
	id: text('id').primaryKey(),
	userId: text('user_id')
		.notNull()
		.references(() => user.id),
	name: text('name'),
	setcode: text('setcode'),
	setname: text('setname'),
	colors: text('colors'),
	mana_cost: text('mana_cost'),
	image_uri: text('image_uri'),
	collector_number: text('collector_number'),
	quantity: integer('quantity')
});

export const collection_cards = sqliteTable(
	'collection_cards',
	{
		collection_id: text('collection_id')
			.notNull()
			.references(() => collections.id),
		card_id: text('card_id')
			.notNull()
			.references(() => cards.id),
		quantity: integer('quantity')
	},
	(table) => ({
		pk: primaryKey(table.collection_id, table.card_id)
	})
);
