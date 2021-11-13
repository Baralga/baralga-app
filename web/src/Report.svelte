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
    <nav class="level ">
      <div class="level-left">
        <div class="level-item report-filter">

          <div class="select">
            <select bind:value={selectedTimespan} on:change={resetTimespan}>
              <option value="week">Week</option>
              <option value="month">Month</option>
              <option value="quarter">Quarter</option>
              <option value="year">Year</option>
            </select>
          </div>

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
    {#each $filteredActivitiesStore as activity}
      <div class="column is-12">
        <div class="card" title={activity.description}>
          <header class="card-header">
            <p class="card-header-title">
              {activity.startTime.format('DD.MM.YYYY')}
            </p>
            <a href="/activities/{activity.id}/edit" class="card-header-icon">
              <span class="icon">
                <i class="fas fa-edit" aria-hidden="true" />
              </span>
            </a>
          </header>
          <div class="card-content">
            <div class="content">
              <div class="columns is-mobile">
                <div class="column is-4">{activity.project.title}</div>
                <div class="column is-5">
                  {activity.startTime.format('HH:mm')} - {activity.endTime.format('HH:mm')}
                </div>
                <div class="column is-3 nowrap">
                  {asFormattedDuration(activity.startTime, activity.endTime)} h
                </div>
              </div>
            </div>
          </div>
          <!--
          <footer class="card-footer">
            <a href="#" class="card-footer-item">Save</a>
          </footer>
          -->
        </div>
      </div>
    {/each}
    <div class="column is-12">
      <div class="card">
        <header class="card-header">
          <div class="card-header-title columns is-mobile">
            <div class="column is-6" />
            <div class="column is-3">Total:</div>
            <div class="column is-3 nowrap">
              {formatDuration($totalDurationStore)} h
            </div>
          </div>
        </header>
        <!--
          <footer class="card-footer">
            <a href="#" class="card-footer-item">Save</a>
          </footer>
          -->
      </div>
    </div>
  {:else}
    <div class="column is-12">
      <span>No activities in this period.</span>
    </div>
  {/if}

</div>
