<script>
  import FileSaver from "file-saver/src/FileSaver";
  import { filteredActivitiesStore, filterStore } from "./stores.js";
  import { asFilterLabel } from "./formatter.js";
  import { createCsv } from "./export_csv.js";

  function exportCsv() {
    let blob = createCsv($filteredActivitiesStore);
    FileSaver.saveAs(
      blob,
      "Activities-" + asFilterLabel($filterStore) + ".csv"
    );
  }
</script>

<button class="button is-link is-hidden-mobile" on:click={exportCsv}>
  <span class="icon">
    <i class="fas fa-file-csv" />
  </span>
  <span>
    Export as
    <abbr title="Comma Separated Values">CSV</abbr>
  </span>
</button>

<button class="button is-link is-hidden-desktop is-hidden-tablet" on:click={exportCsv}>
  <span class="icon">
    <i class="fas fa-file-csv" />
  </span>
</button>
