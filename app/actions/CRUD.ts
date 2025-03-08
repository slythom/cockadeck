'use server'

let deckList: any[] = []; // Liste globale en mémoire
 
export async function createDeck(formData: FormData) {
  const setCode = formData.get('setcode')
  const collectorNumber = formData.get('collectornumber')
  const response = await fetch(`https://api.scryfall.com/cards/${setCode}/${collectorNumber}`);
  const result = await response.json();
  deckList.push(result)
  return result;
}

export async function handleExport() {
    console.log(deckList);
    return deckList;
}