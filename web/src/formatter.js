import moment from "moment/src/moment";

export const formatDuration = (duration) => {
  return (
    duration.hours().toString().padStart(2, "0") +
    ":" +
    duration.minutes().toString().padStart(2, "0")
  );
};

export const asFilterLabel = (filter) => {
  switch (filter.timespan) {
    case "year":
      return filter.from.format("YYYY");
    case "quarter":
      return (
        filter.from.format("Qo") + " Quarter " + filter.from.format("YYYY")
      );
    case "month":
      return filter.from.format("MMMM YYYY");
    case "week":
      return (
        filter.from.format("wo") + " Week " + filter.from.format("MMMM YYYY")
      );
    default:
      return "?";
  }
};

export const asFormattedDuration = (from, to) => {
  let duration = moment.duration(to.diff(from));
  return formatDuration(duration);
};
