<script lang="ts">
  import { parseDate, today } from "@internationalized/date";
  import { Calendar, Popover } from "bits-ui";
  import { CalendarDays, ChevronLeft, ChevronRight } from "@lucide/svelte";

  let { value, onChange }: { value: string; onChange: (date: string) => void } = $props();
  let open = $state(false);
  const timezone = "Asia/Jakarta";
  const currentDate = $derived(open ? today(timezone) : undefined);
  const selected = $derived(value ? parseDate(value) : undefined);
  const formatter = new Intl.DateTimeFormat("id-ID", { dateStyle: "long", timeZone: timezone });
</script>

<Popover.Root bind:open>
  <Popover.Trigger
    aria-label="Tanggal event"
    class="flex w-full items-center gap-2 rounded-lg border border-border bg-muted px-3 py-2 text-left text-sm hover:bg-accent"
  >
    <CalendarDays class="size-4 text-muted-foreground" />
    {selected ? formatter.format(selected.toDate(timezone)) : "Pilih tanggal"}
  </Popover.Trigger>
  <Popover.Portal>
    <Popover.Content align="start" sideOffset={6} class="z-[60] rounded-xl border border-border bg-card p-3 text-foreground shadow-xl">
      <Calendar.Root type="single" value={selected} placeholder={selected ?? currentDate}
        locale="id-ID" weekStartsOn={1} preventDeselect initialFocus calendarLabel="Tanggal event"
        onValueChange={(date) => {
          if (date) onChange(date.toString());
          open = false;
        }}
      >
        {#snippet children({ months, weekdays })}
          <Calendar.Header class="mb-2 flex items-center justify-between gap-2">
            <Calendar.PrevButton aria-label="Bulan sebelumnya" class="rounded p-1.5 hover:bg-accent"><ChevronLeft class="size-4" /></Calendar.PrevButton>
            <Calendar.Heading class="text-sm font-semibold" />
            <Calendar.NextButton aria-label="Bulan berikutnya" class="rounded p-1.5 hover:bg-accent"><ChevronRight class="size-4" /></Calendar.NextButton>
          </Calendar.Header>
          {#each months as month (month.value.toString())}
            <Calendar.Grid class="border-separate border-spacing-0.5">
              <Calendar.GridHead>
                <Calendar.GridRow>
                  {#each weekdays as weekday}
                    <Calendar.HeadCell class="text-center text-xs font-medium text-muted-foreground">{weekday}</Calendar.HeadCell>
                  {/each}
                </Calendar.GridRow>
              </Calendar.GridHead>
              <Calendar.GridBody>
                {#each month.weeks as week (week[0].toString())}
                  <Calendar.GridRow>
                    {#each week as day (day.toString())}
                      <Calendar.Cell date={day} month={month.value}>
                        <Calendar.Day
                          data-past={currentDate && day.compare(currentDate) < 0 ? "" : undefined}
                          class="grid size-9 cursor-pointer place-items-center rounded-md text-sm hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring data-past:opacity-45 data-outside-month:opacity-35 data-today:ring-1 data-today:ring-primary data-selected:bg-primary data-selected:text-primary-foreground data-selected:opacity-100"
                        />
                      </Calendar.Cell>
                    {/each}
                  </Calendar.GridRow>
                {/each}
              </Calendar.GridBody>
            </Calendar.Grid>
          {/each}
        {/snippet}
      </Calendar.Root>
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>
