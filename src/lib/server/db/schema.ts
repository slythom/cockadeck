// import { mysqlTable, serial, int, varchar, datetime } from 'drizzle-orm/mysql-core';

// export const user = mysqlTable('user', {
// 	id: varchar('id', { length: 255 }).primaryKey(),
// 	age: int('age'),
// 	username: varchar('username', { length: 32 }).notNull().unique(),
// 	passwordHash: varchar('password_hash', { length: 255 }).notNull()
// });

// export const session = mysqlTable('session', {
// 	id: varchar('id', { length: 255 }).primaryKey(),
// 	userId: varchar('user_id', { length: 255 })
// 		.notNull()
// 		.references(() => user.id),
// 	expiresAt: datetime('expires_at').notNull()
// });

// export type Session = typeof session.$inferSelect;

// export type User = typeof user.$inferSelect;
import {
	mysqlTable,
	varchar,
	text,
	int,
	timestamp,
	primaryKey,
	unique,
	index
} from "drizzle-orm/mysql-core";
import { relations } from "drizzle-orm";

/* =========================
   USER
========================= */

export const user = mysqlTable("user", {
	id: varchar("id", { length: 36 }).primaryKey(),
	email: varchar("email", { length: 255 }).notNull().unique(),
	passwordHash: varchar("password_hash", { length: 255 }).notNull(),
	createdAt: timestamp("created_at").defaultNow().notNull()
});

export const session = mysqlTable(
	"session",
	{
		id: varchar("id", { length: 255 }).primaryKey(),

		userId: varchar("user_id", { length: 36 })
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),

		expiresAt: timestamp("expires_at").notNull(),

		createdAt: timestamp("created_at").defaultNow().notNull()
	},
	(table) => ({
		userIdx: index("session_user_idx").on(table.userId),
		expiresIdx: index("session_expires_idx").on(table.expiresAt)
	})
);


/* =========================
   CARD (Scryfall)
========================= */

export const card = mysqlTable(
	"card",
	{
		id: int("id").autoincrement().primaryKey(),

		// Scryfall
		scryfallId: varchar("scryfall_id", { length: 36 }).notNull(),
		name: varchar("name", { length: 255 }).notNull(),
		setCode: varchar("set_code", { length: 10 }).notNull(),
		collectorNumber: varchar("collector_number", { length: 20 }).notNull(),

		imageUrl: text("image_url").notNull(),
		oracleText: text("oracle_text"),

		createdAt: timestamp("created_at").defaultNow().notNull()
	},
	(table) => ({
		scryfallIdx: unique().on(table.scryfallId),
		searchIdx: index("card_search_idx").on(
			table.name,
			table.setCode,
			table.collectorNumber
		)
	})
);

/* =========================
   USER COLLECTION
========================= */

export const collectionCard = mysqlTable(
	"collection_card",
	{
		userId: varchar("user_id", { length: 36 })
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),

		cardId: int("card_id")
			.notNull()
			.references(() => card.id, { onDelete: "cascade" }),

		quantity: int("quantity").notNull().default(1),

		createdAt: timestamp("created_at").defaultNow().notNull()
	},
	(table) => ({
		pk: primaryKey({ columns: [table.userId, table.cardId] }),
		cardIdx: index("collection_card_card_idx").on(table.cardId)
	})
);

/* =========================
   DECK
========================= */

export const deck = mysqlTable(
	"deck",
	{
		id: int("id").autoincrement().primaryKey(),

		userId: varchar("user_id", { length: 36 })
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),

		name: varchar("name", { length: 255 }).notNull(),

		createdAt: timestamp("created_at").defaultNow().notNull(),
		updatedAt: timestamp("updated_at").defaultNow().onUpdateNow().notNull()
	},
	(table) => ({
	userIdx: index("deck_user_idx").on(table.userId),
	uniqueName: unique("deck_user_name_unique").on(table.userId, table.name)
})
);

/* =========================
   DECK CARDS
========================= */

export const deckCard = mysqlTable(
	"deck_card",
	{
		deckId: int("deck_id")
			.notNull()
			.references(() => deck.id, { onDelete: "cascade" }),

		cardId: int("card_id")
			.notNull()
			.references(() => card.id, { onDelete: "cascade" }),

		quantity: int("quantity").notNull().default(1)
	},
	(table) => ({
	pk: primaryKey({ columns: [table.deckId, table.cardId] }),
	cardIdx: index("deck_card_card_idx").on(table.cardId)
})
);

/* =========================
   RELATIONS
========================= */

export const userRelations = relations(user, ({ many }) => ({
	decks: many(deck),
	collection: many(collectionCard),
	sessions: many(session)
}));

export const sessionRelations = relations(session, ({ one }) => ({
	user: one(user, {
		fields: [session.userId],
		references: [user.id]
	})
}));

export const cardRelations = relations(card, ({ many }) => ({
	inCollections: many(collectionCard),
	inDecks: many(deckCard)
}));

export const deckRelations = relations(deck, ({ many, one }) => ({
	user: one(user, {
		fields: [deck.userId],
		references: [user.id]
	}),
	cards: many(deckCard)
}));

export const collectionCardRelations = relations(collectionCard, ({ one }) => ({
	user: one(user, {
		fields: [collectionCard.userId],
		references: [user.id]
	}),
	card: one(card, {
		fields: [collectionCard.cardId],
		references: [card.id]
	})
}));

export const deckCardRelations = relations(deckCard, ({ one }) => ({
	deck: one(deck, {
		fields: [deckCard.deckId],
		references: [deck.id]
	}),
	card: one(card, {
		fields: [deckCard.cardId],
		references: [card.id]
	})
}));

