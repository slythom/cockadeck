"use server";
import { Builder } from "xml2js";
import fs from 'node:fs';

let deckList = [] as any; // Liste globale en m�moire

export async function createDeck(formData: FormData) {
  const cardQty = formData.get("cardqty");
  const setCode = formData.get("setcode");
  const collectorNumber = formData.get("collectornumber");
  const response = await fetch(
    `https://api.scryfall.com/cards/${setCode}/${collectorNumber}`
  );
  const result = await response.json();
  const cardName = result.name;
  deckList.push({ cardQty, cardName });
  return deckList;
}

// CREATION XML
export async function handleExport() {
  const xmlEntries = {
    cockatrice_deck: {
      $: {
        version: "1",
      },
      lastLoadedTimestamp: "",
      deckname: "",
      bannerCard: {
        $: {providerId: ""}
      },
      comments: "",
      tags: "",
      zone: {
        $: {
          name: "main",
        },
        card: deckList.map((card: any) => ({
          $: {
            number: card.cardQty,
            name: card.cardName,
          },
        })),
      },
    },
  };

  const builder = new Builder();
  const xml = builder.buildObject(xmlEntries);
  console.log(xml);
  fs.writeFile('exportdeck.cod', xml, err => {
    if (err) {
      console.error(err);
    } else {
      // file written successfully
    }
  });
}