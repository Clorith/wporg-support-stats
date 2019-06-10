# WordPress.org Support Stats
Scrape for support stats at set intervals from WordPress.org support forums.

This is a command-line tool, built using Go. It is built for the major systems (Windows, Mac, Linux), but needs more testers outside of Windows.

The tool will run indefinitely, fetching stats at a configurable rate, until you stop the script.

## Setup and running
To use the scraper, you first need a WordPress.org account with admin rights to a support forum.
After this, rename the `config-sample.json` file to `config.json`, and fill in the apropriate details:
- username : Your WordPress.org username
- password : Your WordPress.org password
- site address : The address to your forums front page
- schedule : How often to run the stats gatherer
  - `hourly` (default) runs on every full hour
  - `daily` runs once a day
  - `weekly` runs once a week

Once the configuration is done, run the `wporg-support-stats` file in your favorite command line interface (holding down <kbd>shift</kbd> and right-clicking will bring up a context-menu with command line interfaces available on Windows systems).

## Output
The gathered data is stored in `output/stats.csv` as a CSV (Comma Separated Values) format, which can be opened by most spreadsheet applications (such as Google Sheets, Microsoft Excel, OpenOffice Calc, and so forth).
