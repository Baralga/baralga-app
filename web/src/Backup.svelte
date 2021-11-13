<script>
  import FileSaver from "file-saver/src/FileSaver";
  import { navigate } from "svelte-routing";
  import { activitiesStore, projectStore, importBackup } from "./stores.js";
  import { readXml, createXml } from "./xml_backup.js";
  import moment from "moment/src/moment";

  let backup;

  function exportXml() {
    let blob = createXml($activitiesStore, $projectStore);
    FileSaver.saveAs(
      blob,
      "backup_" + moment().format("DD.MM.YYYY") + ".baralga.xml"
    );
  }

  function handleFileSelect(event) {
    let selectedFile = event.target.files[0];
    let isValidBackupFile =
      selectedFile.name.endsWith(".xml") && selectedFile.type === "text/xml";
    if (!isValidBackupFile) {
      console.error("Ignore");
      return;
    }

    let reader = new FileReader();
    reader.readAsText(selectedFile);
    reader.onload = () => {
      backup = readXml(reader.result);
      console.info(backup);
    };
  }

  function cancel() {
    backup = null;
  }

  function doImport() {
    importBackup(backup);
    backup = null;
    navigate("/", { replace: true });
  }

  let isFileApiSupported =
    window.File && window.FileReader && window.FileList && window.Blob;
</script>

<h1 class="title is-1">Backup</h1>

<div class="columns is-multiline">
  {#if isFileApiSupported}
    <div class="column is-12">
      <nav class="level ">
        <div class="level-left">
          <div class="level-item">

            <button
              class="button is-link is-hidden-mobile"
              on:click={exportXml}>
              <span class="icon">
                <i class="fas fa-file-code" />
              </span>
              <span>
                Export as
                <abbr title="Extensible Markup Language">XML</abbr>
                backup
              </span>
            </button>

            <button
              class="button is-link is-hidden-desktop is-hidden-tablet"
              on:click={exportXml}>
              <span class="icon">
                <i class="fas fa-file-code" />
              </span>
            </button>
          </div>
          <div class="level-item">

            <div class="file">
              <label class="file-label">
                <input
                  class="file-input"
                  type="file"
                  on:change={handleFileSelect}
                  name="resume" />
                <span class="file-cta">
                  <span class="file-icon">
                    <i class="fas fa-upload" />
                  </span>
                  <span class="file-label">Import from XML backup</span>
                </span>
              </label>
            </div>

          </div>

        </div>

      </nav>
    </div>
    {#if backup}
      <div class="column is-12">
        {#if backup.status == 'ok'}
          <h2 class="title is-2">Confirm Restore</h2>

          <article class="message is-warning">
            <div class="message-body">
              You are about to import
              <strong>{backup.projects.length}</strong>
              projects and
              <strong>{backup.activities.length}</strong>
              activities from your xml backup. Importing that data will
              overwrite and
              <i>all</i>
              your existing projects and activities.
            </div>
          </article>

          <div class="field is-grouped">
            <p class="control">

              <button class="button is-success" on:click={doImport}>
                Yes, import and overwrite all my data
              </button>
              <button class="button" on:click={cancel}>Cancel</button>
            </p>
          </div>
        {:else}
          <h2 class="title is-2">Reading Backup failed</h2>

          <article class="message is-danger">
            <div class="message-body">
              The provided backup xml file could not be processed. Please check
              your file and try again.
            </div>
          </article>
        {/if}

      </div>
    {/if}
  {:else}
    <div class="column is-12">
      <article class="message is-warning">
        <div class="message-body">
          Your browser does not support the HTML5 file api. Please upgrade your
          browser.
        </div>
      </article>
    </div>
  {/if}

</div>
