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

import filtersCatalog from '../../helpers/filters/filters';
import { FilteringData } from '../../initialState';

interface DnsBlocklistProps {
    getFilteringStatus: (...args: unknown[]) => unknown;
    filtering: FilteringData;
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

class DnsBlocklist extends Component<DnsBlocklistProps> {
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
        const { modalFilterUrl, modalType } = this.props.filtering;

        switch (modalType) {
            case MODAL_TYPE.EDIT_FILTERS:
                this.props.editFilter(modalFilterUrl, values);
                break;
            case MODAL_TYPE.ADD_FILTERS: {
                const { name, url } = values;

                this.props.addFilter(url, name);
                break;
            }
            case MODAL_TYPE.ADD_FILTERS_BULK: {
                const { bulkUrls } = values;
                const entries = parseBulkFiltersInput(bulkUrls);

                if (entries.length > 0) {
                    await this.props.addFiltersBulk(entries);
                }
                break;
            }
            case MODAL_TYPE.CHOOSE_FILTERING_LIST: {
                const changedValues = Object.entries(values)?.reduce((acc: any, [key, value]) => {
                    if (value && key in filtersCatalog.filters) {
                        acc[key] = value;
                    }
                    return acc;
                }, {});

                Object.keys(changedValues).forEach((fieldName) => {
                    // filterId is actually in the field name

                    const { source, name } = filtersCatalog.filters[fieldName];

                    this.props.addFilter(source, name);
                });
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
        this.props.toggleFilterStatus(url, data);
    };

    handleRefresh = () => {
        this.props.refreshFilters({ whitelist: false });
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
        const { filters } = this.props.filtering;
        this.setState({ selectedUrls: new Set(filters.map((f: any) => f.url)) });
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
        this.setState({ isBulkConfirmOpen: false, selectedUrls: new Set<string>() });
        await this.props.removeFiltersBulk(urls);
    };

    render() {
        const {
            t,

            toggleFilteringModal,

            addFilter,

            filtering: {
                filters,
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
        const currentFilterData = getCurrentFilter(modalFilterUrl, filters);
        const enabledFiltersRulesCount = filters.reduce((acc: number, filter: Filter) => {
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

        return (
            <>
                <PageTitle title={t('dns_blocklists')} subtitle={t('dns_blocklists_desc')} />

                <div className="content">
                    <div className="row">
                        <div className="col-md-12">
                            <Card subtitle={t('filters_and_hosts_hint')}>
                                <Summary
                                    label={t('domains_on_blocklists')}
                                    hint={t('domains_on_blocklists_hint')}
                                    total={enabledFiltersRulesCount}
                                />

                                <Table
                                    filters={filters}
                                    loading={loading}
                                    processingConfigFilter={processingConfigFilter}
                                    toggleFilteringModal={toggleFilteringModal}
                                    handleDelete={this.handleDelete}
                                    toggleFilter={this.toggleFilter}
                                    selectedUrls={this.state.selectedUrls}
                                    onToggleSelect={this.handleToggleSelect}
                                    onSelectAll={this.handleSelectAll}
                                    onDeselectAll={this.handleDeselectAll}
                                />

                                <Actions
                                    handleAdd={this.openSelectTypeModal}
                                    handleRefresh={this.handleRefresh}
                                    processingRefreshFilters={processingRefreshFilters}
                                    selectedCount={this.state.selectedUrls.size}
                                    onDeleteSelected={this.handleDeleteSelected}
                                    processingRemoveFilter={processingRemoveFilter}
                                />
                            </Card>
                        </div>
                    </div>
                </div>

                <Modal
                    filtersCatalog={filtersCatalog}
                    filters={filters}
                    isOpen={isModalOpen}
                    toggleFilteringModal={toggleFilteringModal}
                    addFilter={addFilter}
                    isFilterAdded={isFilterAdded}
                    processingAddFilter={processingAddFilter}
                    processingConfigFilter={processingConfigFilter}
                    handleSubmit={this.handleSubmit}
                    modalType={modalType}
                    currentFilterData={currentFilterData}
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
                                    this.props.removeFilter(this.state.deleteUrl);
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

export default withTranslation()(DnsBlocklist);
