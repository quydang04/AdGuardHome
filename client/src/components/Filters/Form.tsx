import React from 'react';
import { useForm, Controller, FormProvider } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { validateBulkFilterUrls, validatePath, validateRequiredValue } from '../../helpers/validators';

import { MODAL_OPEN_TIMEOUT, MODAL_TYPE } from '../../helpers/constants';
import filtersCatalog from '../../helpers/filters/filters';
import { FiltersList } from './FiltersList';
import { Input } from '../ui/Controls/Input';
import { Textarea } from '../ui/Controls/Textarea';

type FormValues = {
    enabled: boolean;
    name: string;
    url: string;
    bulkUrls?: string;
};

const defaultValues: FormValues = {
    enabled: true,
    name: '',
    url: '',
    bulkUrls: '',
};

type Props = {
    closeModal: () => void;
    onSubmit: (values: FormValues) => void;
    processingAddFilter: boolean;
    processingConfigFilter: boolean;
    whitelist?: boolean;
    modalType: string;
    toggleFilteringModal: ({ type }: { type?: keyof typeof MODAL_TYPE }) => void;
    selectedSources?: Record<string, boolean>;
    initialValues?: FormValues;
    canChooseFromCatalog?: boolean;
};

export const Form = ({
    closeModal,
    processingAddFilter,
    processingConfigFilter,
    whitelist,
    modalType,
    toggleFilteringModal,
    selectedSources,
    onSubmit,
    initialValues,
    canChooseFromCatalog = true,
}: Props) => {
    const { t } = useTranslation();

    const methods = useForm({
        defaultValues: {
            ...defaultValues,
            ...initialValues,
        },
        mode: 'onBlur',
    });
    const { handleSubmit, control } = methods;

    const openModal = (modalType: keyof typeof MODAL_TYPE, timeout = MODAL_OPEN_TIMEOUT) => {
        toggleFilteringModal(undefined);
        setTimeout(() => toggleFilteringModal({ type: modalType }), timeout);
    };

    const openFilteringListModal = () => {
        if (!canChooseFromCatalog) {
            return;
        }
        openModal('CHOOSE_FILTERING_LIST');
    };

    const openAddFiltersModal = () => openModal('ADD_FILTERS');

    const openAddFiltersBulkModal = () => openModal('ADD_FILTERS_BULK');

    const showSingleFilterFields =
        modalType !== MODAL_TYPE.CHOOSE_FILTERING_LIST &&
        modalType !== MODAL_TYPE.SELECT_MODAL_TYPE &&
        modalType !== MODAL_TYPE.ADD_FILTERS_BULK;

    const showBulkFilterFields = modalType === MODAL_TYPE.ADD_FILTERS_BULK;

    const bulkButtonKey = whitelist ? 'add_multiple_allowlists' : 'add_multiple_blocklists';
    const bulkHintKey = whitelist ? 'bulk_allowlist_hint' : 'bulk_blocklist_hint';
    const bulkPlaceholderKey = whitelist ? 'bulk_allowlist_placeholder' : 'bulk_blocklist_placeholder';

    return (
        <FormProvider {...methods}>
            <form onSubmit={handleSubmit(onSubmit)}>
                <div className="modal-body modal-body--filters">
                    {modalType === MODAL_TYPE.SELECT_MODAL_TYPE && (
                        <div className="d-flex justify-content-around flex-wrap">
                            {canChooseFromCatalog && (
                                <button
                                    type="button"
                                    onClick={openFilteringListModal}
                                    className="btn btn-success btn-standard mr-2 btn-large">
                                    {t('choose_from_list')}
                                </button>
                            )}

                            <button
                                type="button"
                                onClick={openAddFiltersModal}
                                className="btn btn-primary btn-standard">
                                {t('add_custom_list')}
                            </button>

                            <button
                                type="button"
                                onClick={openAddFiltersBulkModal}
                                className="btn btn-secondary btn-standard mt-2">
                                {t(bulkButtonKey)}
                            </button>
                        </div>
                    )}
                    {modalType === MODAL_TYPE.CHOOSE_FILTERING_LIST && (
                        <FiltersList
                            categories={filtersCatalog.categories}
                            filters={filtersCatalog.filters}
                            selectedSources={selectedSources}
                        />
                    )}
                    {showSingleFilterFields && (
                        <>
                            <div className="form__group">
                                <Controller
                                    name="name"
                                    control={control}
                                    render={({ field, fieldState }) => (
                                        <Input
                                            {...field}
                                            type="text"
                                            data-testid="filters_name"
                                            placeholder={t('enter_name_hint')}
                                            error={fieldState.error?.message}
                                            trimOnBlur
                                        />
                                    )}
                                />
                            </div>

                            <div className="form__group">
                                <Controller
                                    name="url"
                                    control={control}
                                    rules={{ validate: { validateRequiredValue, validatePath } }}
                                    render={({ field, fieldState }) => (
                                        <Input
                                            {...field}
                                            type="text"
                                            data-testid="filters_url"
                                            placeholder={t('enter_url_or_path_hint')}
                                            error={fieldState.error?.message}
                                            trimOnBlur
                                        />
                                    )}
                                />
                            </div>

                            <div className="form__description">
                                {whitelist ? t('enter_valid_allowlist') : t('enter_valid_blocklist')}
                            </div>
                        </>
                    )}

                    {showBulkFilterFields && (
                        <div className="form__group">
                            <Controller
                                name="bulkUrls"
                                control={control}
                                rules={{ validate: { validateRequiredValue, validateBulkFilterUrls } }}
                                render={({ field, fieldState }) => (
                                    <Textarea
                                        {...field}
                                        rows={8}
                                        data-testid="filters_bulk_urls"
                                        placeholder={t(bulkPlaceholderKey)}
                                        error={fieldState.error?.message}
                                        trimOnBlur
                                    />
                                )}
                            />

                            <div className="form__description">{t(bulkHintKey)}</div>
                        </div>
                    )}
                </div>

                <div className="modal-footer">
                    <button type="button" className="btn btn-secondary" onClick={closeModal}>
                        {t('cancel_btn')}
                    </button>

                    {modalType !== MODAL_TYPE.SELECT_MODAL_TYPE && (
                        <button
                            type="submit"
                            data-testid="filters_save"
                            className="btn btn-success"
                            disabled={processingAddFilter || processingConfigFilter}>
                            {t('save_btn')}
                        </button>
                    )}
                </div>
            </form>
        </FormProvider>
    );
};
