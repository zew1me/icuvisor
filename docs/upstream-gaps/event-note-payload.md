# intervals.icu NOTE event create payload

Live probing for TP-075 showed that intervals.icu bulk event creates accept NOTE events only when `start_date_local` is a local date-time string such as `YYYY-MM-DDT00:00:00`. The same payload with a date-only `YYYY-MM-DD` value is rejected by upstream with HTTP 422. This differs from icuvisor's public tool argument, which remains a date-only athlete-local `date`; the intervals client is responsible for adding the midnight time component before POSTing NOTE creates.

The probed minimal accepted NOTE create payload was `category: "NOTE"`, a non-empty `name`, and `start_date_local: "YYYY-MM-DDT00:00:00"`. `type` may be omitted or empty for NOTE events, and `description` is optional when `name` is present. Description-only NOTE creates were rejected with `Name is required`, and category casing is strict: `NOTE` worked while `Note` and `note` were rejected.
