import { computed, ref } from "vue";
import type { ComputedRef, Ref } from "vue";
import type { DataTablePageEvent } from "primevue/datatable";

export interface CursorPager {
    pageSize: Ref<number>;
    nextCursor: Ref<string | null>;
    cursor: ComputedRef<string | null>;
    first: ComputedRef<number>;
    total: (rows: number) => number;
    turn: (event: DataTablePageEvent) => boolean;
    reset: () => void;
}

export function useCursorPager(firstPageSize: number): CursorPager {
    const pageSize = ref<number>(firstPageSize);
    const cursors = ref<(string | null)[]>([null]);
    const nextCursor = ref<string | null>(null);

    const cursor = computed<string | null>(() => cursors.value.at(-1) ?? null);
    const first = computed<number>(() => (cursors.value.length - 1) * pageSize.value);

    function total(rows: number): number {
        return first.value + rows + Number(nextCursor.value !== null);
    }

    // Returns true when the page must load.
    function turn(event: DataTablePageEvent): boolean {
        if (event.rows !== pageSize.value) {
            pageSize.value = event.rows;
            reset();
            return true;
        }

        const page = cursors.value.length - 1;
        if (event.page === 0) cursors.value = [null];
        else if (event.page === page - 1) cursors.value.pop();
        else if (event.page === page + 1 && nextCursor.value !== null) cursors.value.push(nextCursor.value);
        else return false;
        nextCursor.value = null;
        return true;
    }

    function reset(): void {
        cursors.value = [null];
    }

    return { pageSize, nextCursor, cursor, first, total, turn, reset };
}
