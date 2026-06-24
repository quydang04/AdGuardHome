import React, { Component } from 'react';
import { withTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Modal from './Modal';
import Actions from './Actions';
import Table from './Table';
import Summary from './Summary';

import { MODAL_TYPE } from '../../helpers/constants';
import ReactModal from 'react-modal';

import { Filter, getCurrentFilter } from '../../helpers/helpers';
import { parseBulkFiltersInput } from '../../helpers/filteringBulk';

interface DnsAllowlistProps {
    getFilteringStatus: (...args: unknown[]) => unknown;
    filtering: {
        modalType: string;
        modalFilterUrl: string;
        isModalOpen: boolean;
        isFilterAdded: boolean;
        processingRefreshFilters: boolean;
        processingRemoveFilter: boolean;
        processingAddFilter: boolean;
        processingConfigFilter: boolean;
        processingFilters: boolean;
        whitelistFilters: any[];
    };
    removeFilter: (...args: unknown[]) => unknown;
    removeFiltersBulk: (...args: unknown[]) => unknown;
    toggleFilterStatus: (...args: unknown[]) => unknown;
    addFilter: (...args: unknown[]) => unknown;
    addFiltersBulk: (...args: unknown[]) => unknown;
    toggleFilteringModal: (...args: unknown[]) => unknown;
    handleRulesChange: (...args: unknown[]) => unknown;
    refreshFilters: (...args: unknown[]) => unknown;
    editFilter: (...args: unknown[]) => unknown;
    t: (...args: unknown[]) => string;
}

const canChooseFromCatalog = false;

class DnsAllowlist extends Component<DnsAllowlistProps> {
    state = {
        isConfirmOpen: false,
        deleteUrl: '',
        selectedUrls: new Set<string>(),
        isBulkConfirmOpen: false,
    };

    componentDidMount() {
        this.props.getFilteringStatus();
    }

    handleSubmit = async (values: any) => {
        const { filtering } = this.props;
        const whitelist = true;

        switch (filtering.modalType) {
            case MODAL_TYPE.EDIT_FILTERS:
                this.props.editFilter(filtering.modalFilterUrl, values, whitelist);
                break;
            case MODAL_TYPE.ADD_FILTERS: {
                const { name, url } = values;
                this.props.addFilter(url, name, whitelist);
                break;
            }
            case MODAL_TYPE.ADD_FILTERS_BULK: {
                const { bulkUrls } = values;
                const entries = parseBulkFiltersInput(bulkUrls);

                if (entries.length > 0) {
                    await this.props.addFiltersBulk(entries, whitelist);
                }
                break;
            }
            default:
                break;
        }
    };

    handleDelete = (url: any) => {
        this.setState({
            isConfirmOpen: true,
            deleteUrl: url,
        });
    };

    toggleFilter = (url: any, data: any) => {
        const whitelist = true;

        this.props.toggleFilterStatus(url, data, whitelist);
    };

    handleRefresh = () => {
        this.props.refreshFilters({ whitelist: true });
    };

    openSelectTypeModal = () => {
        this.props.toggleFilteringModal({ type: MODAL_TYPE.SELECT_MODAL_TYPE });
    };

    handleToggleSelect = (url: string) => {
        this.setState((prevState: any) => {
            const next = new Set(prevState.selectedUrls);
            if (next.has(url)) {
                next.delete(url);
            } else {
                next.add(url);
            }
            return { selectedUrls: next };
        });
    };

    handleSelectAll = () => {
        const { whitelistFilters } = this.props.filtering;
        this.setState({ selectedUrls: new Set(whitelistFilters.map((f: any) => f.url)) });
    };

    handleDeselectAll = () => {
        this.setState({ selectedUrls: new Set<string>() });
    };

    handleDeleteSelected = () => {
        if (this.state.selectedUrls.size > 0) {
            this.setState({ isBulkConfirmOpen: true });
        }
    };

    confirmBulkDelete = async () => {
        const urls = Array.from(this.state.selectedUrls);
        const whitelist = true;
        this.setState({ isBulkConfirmOpen: false, selectedUrls: new Set<string>() });
        await this.props.removeFiltersBulk(urls, whitelist);
    };

    render() {
        const {
            t,
            toggleFilteringModal,
            addFilter,
            filtering: {
                whitelistFilters,
                isModalOpen,
                isFilterAdded,
                processingRefreshFilters,
                processingRemoveFilter,
                processingAddFilter,
                processingConfigFilter,
                processingFilters,
                modalType,
                modalFilterUrl,
            },
        } = this.props;
        const currentFilterData = getCurrentFilter(modalFilterUrl, whitelistFilters);
        const enabledAllowlistRulesCount = whitelistFilters.reduce((acc: number, filter: Filter) => {
            if (!filter?.enabled) {
                return acc;
            }

            return acc + (filter?.rulesCount || 0);
        }, 0);
        const loading =
            processingConfigFilter ||
            processingFilters ||
            processingAddFilter ||
            processingRemoveFilter ||
            processingRefreshFilters;

        const whitelist = true;

        return (
            <>
                <PageTitle title={t('dns_allowlists')} subtitle={t('dns_allowlists_desc')} />
                <div className="content">
                    <div className="row">
                        <div className="col-md-12">
                            <Card subtitle={t('filters_and_hosts_hint')}>
                                <Summary
                                    label={t('domains_on_allowlists')}
                                    hint={t('domains_on_allowlists_hint')}
                                    total={enabledAllowlistRulesCount}
                                />

                                <Table
                                    filters={whitelistFilters}
                                    loading={loading}
                                    processingConfigFilter={processingConfigFilter}
                                    toggleFilteringModal={toggleFilteringModal}
                                    handleDelete={this.handleDelete}
                                    toggleFilter={this.toggleFilter}
                                    whitelist={whitelist}
                                    selectedUrls={this.state.selectedUrls}
                                    onToggleSelect={this.handleToggleSelect}
                                    onSelectAll={this.handleSelectAll}
                                    onDeselectAll={this.handleDeselectAll}
                                />

                                <Actions
                                    handleAdd={this.openSelectTypeModal}
                                    handleRefresh={this.handleRefresh}
                                    processingRefreshFilters={processingRefreshFilters}
                                    whitelist={whitelist}
                                    selectedCount={this.state.selectedUrls.size}
                                    onDeleteSelected={this.handleDeleteSelected}
                                    processingRemoveFilter={processingRemoveFilter}
                                />
                            </Card>
                        </div>
                    </div>
                </div>

                <Modal
                    filters={whitelistFilters}
                    isOpen={isModalOpen}
                    toggleFilteringModal={toggleFilteringModal}
                    addFilter={addFilter}
                    isFilterAdded={isFilterAdded}
                    processingAddFilter={processingAddFilter}
                    processingConfigFilter={processingConfigFilter}
                    handleSubmit={this.handleSubmit}
                    modalType={modalType}
                    currentFilterData={currentFilterData}
                    whitelist={whitelist}
                    canChooseFromCatalog={canChooseFromCatalog}
                />

                <ReactModal
                    className="Modal__Bootstrap modal-dialog modal-dialog-centered"
                    closeTimeoutMS={0}
                    isOpen={this.state.isConfirmOpen}
                    onRequestClose={() => this.setState({ isConfirmOpen: false })}>
                    <div className="modal-content">
                        <div className="modal-header">
                            <h4 className="modal-title">{t('delete_table_action')}</h4>
                            <button
                                type="button"
                                className="close"
                                onClick={() => this.setState({ isConfirmOpen: false })}>
                                <span className="sr-only">Close</span>
                            </button>
                        </div>
                        <div className="modal-body">
                            <p className="mb-0">{t('list_confirm_delete')}</p>
                        </div>
                        <div className="modal-footer justify-content-end gap-2" style={{ display: 'flex' }}>
                            <button
                                type="button"
                                className="btn btn-secondary mr-2"
                                onClick={() => this.setState({ isConfirmOpen: false })}>
                                {t('cancel_btn')}
                            </button>
                            <button
                                type="button"
                                className="btn btn-danger"
                                onClick={() => {
                                    this.props.removeFilter(this.state.deleteUrl, whitelist);
                                    this.setState({ isConfirmOpen: false, deleteUrl: '' });
                                }}>
                                {t('delete_table_action')}
                            </button>
                        </div>
                    </div>
                </ReactModal>

                <ReactModal
                    className="Modal__Bootstrap modal-dialog modal-dialog-centered"
                    closeTimeoutMS={0}
                    isOpen={this.state.isBulkConfirmOpen}
                    onRequestClose={() => this.setState({ isBulkConfirmOpen: false })}>
                    <div className="modal-content">
                        <div className="modal-header">
                            <h4 className="modal-title">{t('delete_table_action')}</h4>
                            <button
                                type="button"
                                className="close"
                                onClick={() => this.setState({ isBulkConfirmOpen: false })}>
                                <span className="sr-only">Close</span>
                            </button>
                        </div>
                        <div className="modal-body">
                            <p className="mb-0">
                                {t('list_confirm_delete_selected', { count: this.state.selectedUrls.size })}
                            </p>
                        </div>
                        <div className="modal-footer justify-content-end gap-2" style={{ display: 'flex' }}>
                            <button
                                type="button"
                                className="btn btn-secondary mr-2"
                                onClick={() => this.setState({ isBulkConfirmOpen: false })}>
                                {t('cancel_btn')}
                            </button>
                            <button
                                type="button"
                                className="btn btn-danger"
                                onClick={this.confirmBulkDelete}>
                                {t('delete_table_action')}
                            </button>
                        </div>
                    </div>
                </ReactModal>
            </>
        );
    }
}

export default withTranslation()(DnsAllowlist);
