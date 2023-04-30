# Go Routine Creator

> **NOTE**: THIS PROJECT HAS AN EXTREMELY BAD TITLE, I JUST THOUGHT IT WAS FUNNY.

This project does not create go routines. If that's what you came for, you've been misled.

This tool helps you create weekly routine and export it to Google Calendar. You can specify a list of tasks, events, and other activities you want to include in your routine, along with their start times and durations.

To use the routine creator with a service account, you'll need to follow these steps:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/) and select your project.
2. Click on the hamburger menu in the top-left corner, and select "IAM & Admin" > "Service accounts".
3. Click on the "Create service account" button and enter a name and description for the service account.
4. Click on the "Create" button to create the service account.
5. After the service account is created, click on the "Add Key" button and select "JSON" as the key type. This will download a JSON file containing the service account credentials.
6. Rename the JSON file to `credentials.json` and save it in the root directory of your application.
7. Grant the service account the necessary permissions to access the Google Calendar API. You can do this by adding the service account email address (found in the JSON file) as a user with appropriate access to the calendar in question.
8. Create a file, `schedule.csv`, with your desired routine. Use `schedule.csv.sample` to start. Each row in the CSV file represents a single task or event, and includes the following columns:
   - Title: The title or name of the task/event
   - Start Time: The start time of the task/event (in the format HH:MM:SS)
   - End Time: The end time of the task/event (in the format HH:MM:SS)
   - Recurrence: (optional) A recurrence rule for the task/event, in the RFC 5545 RRULE format
9. Install the required Go dependencies by running `go mod tidy`.
10. Update the `calendarId` constant in `create-routine.go`
11. Run the `create-routine.go` script to generate a list of events based on your schedule. The script will use the service account credentials to authenticate your application and create events on your behalf.

Enjoy your new routine!

