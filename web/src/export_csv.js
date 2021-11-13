import { asFormattedDuration } from "./formatter.js";

const CSV_HEADER = [
  "Date",
  "Start",
  "End",
  "Duration",
  "Project",
  "Description",
]
  .join(";")
  .concat("\n");

export const createCsv = (activities) => {
  let activitiesRows = activities.map((activity) => {
    return [
      activity.startTime.format("DD.MM.YYYY"),
      activity.startTime.format("HH:mm"),
      activity.endTime.format("HH:mm"),
      asFormattedDuration(activity.startTime, activity.endTime),
      activity.project.name,
      activity.description,
    ]
      .join(";")
      .concat("\n");
  });

  let blob = new Blob([CSV_HEADER, ...activitiesRows], {
    type: "text/csv;charset=utf-8",
  });

  return blob;
};
