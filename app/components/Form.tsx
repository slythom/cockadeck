import { createDeck, handleExport } from "@/app/actions/CRUD";

export default async function Form() {
  
  return (
    <div>
      <form action={createDeck}>
        <input type="text" name="cardqty" id="cardqty" className="outline" />
        <input type="text" name="setcode" id="setcode" className="outline" />
        <input type="text" name="collectornumber" id="collectornumber" className="outline" />
        <button type="submit">Ajouter</button>
      </form>

      <button onClick={handleExport}>Exporter la Liste</button>

    </div>
  );
}
