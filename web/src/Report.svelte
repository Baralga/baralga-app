<script>
  import moment from "moment/src/moment";
  import {
    filterStore,
    filteredActivitiesStore,
    totalDurationStore,
    applyFilter
  } from "./stores.js";
  import {
    formatDuration,
    asFormattedDuration,
    asFilterLabel
  } from "./formatter.js";
  import { onMount } from "svelte";

  let selectedTimespan = "year";
  let filterLabel = "";

  let timespanFrom;
  let timespanTo;

  function updateFilter() {
    let filter = {
      timespan: selectedTimespan,
      from: timespanFrom,
      to: timespanTo
    };

    filterLabel = asFilterLabel(filter);

    applyFilter(filter);
  }

  function resetTimespan() {
    switch (selectedTimespan) {
      case "year":
        timespanFrom = moment().startOf("year");
        timespanTo = moment().endOf("year");
        break;
      case "quarter":
        timespanFrom = moment().startOf("quarter");
        timespanTo = moment().endOf("quarter");
        break;
      case "month":
        timespanFrom = moment().startOf("month");
        timespanTo = moment().endOf("month");
        break;
      case "week":
        timespanFrom = moment().startOf("week");
        timespanTo = moment().endOf("week");
        break;
      case "day":
        timespanFrom = moment().startOf("day");
        timespanTo = moment().endOf("day");
        break;
    }

    updateFilter();
  }

  function moveTimespan(diffToMove) {
    switch (selectedTimespan) {
      case "year":
        timespanFrom = timespanFrom.add(diffToMove, "years");
        timespanTo = timespanTo.add(diffToMove, "years");
        break;
      case "quarter":
        timespanFrom = timespanFrom.add(diffToMove, "quarters");
        timespanTo = timespanTo.add(diffToMove, "quarters");
        break;
      case "month":
        timespanFrom = timespanFrom.add(diffToMove, "months");
        timespanTo = timespanTo.add(diffToMove, "months");
        break;
      case "week":
        timespanFrom = timespanFrom.add(diffToMove, "weeks");
        timespanTo = timespanTo.add(diffToMove, "weeks");
        break;
      case "day":
        timespanFrom = timespanFrom.add(diffToMove, "days");
        timespanTo = timespanTo.add(diffToMove, "days");
        break;
    }
    updateFilter();
  }

  function previousTimespan() {
    moveTimespan(-1);
  }

  function nextTimespan() {
    moveTimespan(1);
  }

  function initFilter() {
    let storedFilter = $filterStore;

    selectedTimespan = storedFilter.timespan;
    timespanFrom = storedFilter.from;
    timespanTo = storedFilter.to;

    updateFilter();
  }

  onMount(function() {
    initFilter();
  });
</script>

<style>
  .nowrap {
    white-space: nowrap;
  }

  .report-filter > div {
    margin-right: 0.5em;
  }
</style>

<div class="columns is-multiline">

  <div class="column is-12">
    <nav class="level">
      <div class="level-left">
        <div class="level-item report-filter">

          <div class="select">
            <select bind:value={selectedTimespan} on:change={resetTimespan}>
              <option value="day">Day</option>
              <option value="week">Week</option>
              <option value="month">Month</option>
              <option value="quarter">Quarter</option>
              <option value="year">Year</option>
            </select>
          </div>

          <div class="buttons has-addons">
            <div class="button" on:click={previousTimespan}>
              <span class="icon is-medium">
                <i class="fa fa-angle-left" />
              </span>
            </div>
            <div class="button" on:click={resetTimespan}>
              <span class="icon is-medium">
                <i class="fa fa-home" />
              </span>
            </div>
            <div class="button" on:click={nextTimespan}>
              <span class="icon is-medium">
                <i class="fa fa-angle-right" />
              </span>
            </div>
          </div>
          
        </div>

      </div>

      <div class="level-right">
        <fieldset disabled>
          <div class="control">
            <input
              class="input"
              type="text"
              placeholder=""
              bind:value={filterLabel} />
          </div>
        </fieldset>
      </div>
    </nav>
  </div>

  {#if $filteredActivitiesStore.length > 0}
   <div class="list column is-12">
    {#each $filteredActivitiesStore as activity}
      <div class="list-item">

        <div class="list-item-content">
          <div class="list-item-title">{activity.startTime.format('DD.MM.YYYY')}</div>
          <div class="list-item-description">

              <div class="content is-size-6 is-size-9-mobile">
                <div class="columns is-mobile">
                <div class="column is-4">{activity.project.title}</div>
                <div class="column is-5">
                  {activity.startTime.format('HH:mm')} - {activity.endTime.format('HH:mm')}
                </div>
                <div class="column is-3 nowrap" title="{activity.startTime.format('HH:mm')} - {activity.endTime.format('HH:mm')}">
                  {activity.duration.formatted}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="list-item-controls">
          <div class="buttons">
            
            <a href="/activities/{activity.id}/edit" class="button is-dark is-inverted">
              <span class="icon">
                <i class="fas fa-pencil-alt"></i>
              </span>
            </a>

          </div>
        </div>
      </div>
      {/each}
    </div>
  {:else}
    <div class="column is-12">
      <span>No activities in this period.</span>
    </div>
  {/if}

</div>
