<script>
  import { filterStore, searchParamsForFilter } from "./stores.js";
</script>

<a class="button is-link is-hidden-mobile" href={"/api/activities?contentType=text/csv&"+ searchParamsForFilter($filterStore).toString()}>
  <span class="icon">
    <i class="fas fa-file-csv" />
  </span>
  <span>
    Export as
    <abbr title="Comma Separated Values">CSV</abbr>
  </span>
</a>

<a class="button is-link is-hidden-desktop is-hidden-tablet" href={"/api/activities?contentType=text/csv&"+ searchParamsForFilter($filterStore).toString()}>
  <span class="icon">
    <i class="fas fa-file-csv" />
  </span>
</a>
