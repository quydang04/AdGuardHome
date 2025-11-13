import React from 'react';
import { ReactNode } from 'react';
import clsx from 'clsx';

export type AccordionItemProps = {
    id: string;
    title: string;
    children: ReactNode;
    defaultOpen?: boolean;
    className?: string;
    disabled?: boolean;
};

type Props = AccordionItemProps & {
  isOpen: boolean;
  onToggle: () => void;
};

export const AccordionItem = (props: Props) => {
    const {
        id,
        title,
        children,
        isOpen,
        onToggle,
        disabled,
        className = '',
    } = props;
    return (
        <section className={clsx('accordion-item', className)} data-testid={`accordion-item-${id}`}>
            <header className="accordion-item__header">
                <div className="accordion-item__toggle-wrapper">
                    <button
                        type="button"
                        className={clsx('accordion-item__toggle', {
                            'accordion-item__toggle--open': isOpen
                        })}
                        onClick={onToggle}
                        aria-expanded={isOpen}
                        aria-controls={`accordion-content-${id}`}
                        aria-disabled={disabled}
                        disabled={disabled}
                    >
                        <span className="accordion-item__icon" aria-hidden="true">
                            <svg width="24" height="24" viewBox="0 0 24 24">
                                <use xlinkHref="#chevron-down" />
                            </svg>
                        </span>
                        <h3 className="accordion-item__title">{title}</h3>
                    </button>
                </div>
            </header>

            <div
                id={`accordion-content-${id}`}
                className={clsx('accordion-item__content', {
                    'accordion-item__content--open': isOpen
                })}
                aria-hidden={!isOpen}
            >
                <div className="accordion-item__content-inner">
                    {children}
                </div>
            </div>
        </section>
    );
};
