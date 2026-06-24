import React from 'react';
import { withTranslation, Trans } from 'react-i18next';

interface ActionsProps {
    handleAdd: (...args: unknown[]) => unknown;
    handleRefresh: (...args: unknown[]) => unknown;
    processingRefreshFilters: boolean;
    whitelist?: boolean;
    // Multi-select delete
    selectedCount?: number;
    onDeleteSelected?: () => void;
    processingRemoveFilter?: boolean;
}

const Actions = ({
    handleAdd,
    handleRefresh,
    processingRefreshFilters,
    whitelist,
    selectedCount = 0,
    onDeleteSelected,
    processingRemoveFilter = false,
}: ActionsProps) => (
    <div className="card-actions">
        <button className="btn btn-success btn-standard mr-2 btn-large mb-2" type="button" onClick={handleAdd}>
            {whitelist ? <Trans>add_allowlist</Trans> : <Trans>add_blocklist</Trans>}
        </button>

        <button
            className="btn btn-primary btn-standard mb-2 mr-2"
            type="button"
            onClick={handleRefresh}
            disabled={processingRefreshFilters}>
            <Trans>check_updates_btn</Trans>
        </button>

        {selectedCount > 0 && onDeleteSelected && (
            <button
                className="btn btn-danger btn-standard mb-2"
                type="button"
                onClick={onDeleteSelected}
                disabled={processingRemoveFilter}>
                <Trans>delete_table_action</Trans>
                {' '}
                ({selectedCount})
            </button>
        )}
    </div>
);

export default withTranslation()(Actions);
