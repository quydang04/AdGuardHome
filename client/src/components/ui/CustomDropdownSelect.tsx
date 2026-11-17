import React, { useEffect, useRef, useState } from 'react';
import classNames from 'classnames';

import './CustomDropdownSelect.css';

export interface CustomDropdownOption<T extends string = string> {
    value: T;
    label: string;
}

interface CustomDropdownSelectProps<T extends string = string> {
    id?: string;
    value: T;
    options: CustomDropdownOption<T>[];
    onChange: (value: T) => void;
    className?: string;
    disabled?: boolean;
}

export const CustomDropdownSelect = <T extends string = string>({
    id,
    value,
    options,
    onChange,
    className,
    disabled = false,
}: CustomDropdownSelectProps<T>) => {
    const [isOpen, setIsOpen] = useState(false);
    const wrapRef = useRef<HTMLDivElement>(null);

    const activeOption = options.find((opt) => opt.value === value) || options[0];

    useEffect(() => {
        if (!isOpen) {
            return undefined;
        }

        const handlePointerDown = (e: Event) => {
            if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
                setIsOpen(false);
            }
        };

        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                setIsOpen(false);
            }
        };

        document.addEventListener('pointerdown', handlePointerDown);
        document.addEventListener('keydown', handleKeyDown);

        return () => {
            document.removeEventListener('pointerdown', handlePointerDown);
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, [isOpen]);

    return (
        <div ref={wrapRef} className={classNames('custom-dd-select', className)}>
            <button
                id={id}
                type="button"
                disabled={disabled}
                aria-haspopup="listbox"
                aria-expanded={isOpen}
                className={classNames('custom-dd-select__trigger', {
                    'custom-dd-select__trigger--open': isOpen,
                })}
                onClick={() => setIsOpen((prev) => !prev)}>
                <span className="custom-dd-select__value">{activeOption?.label || value}</span>
                <svg
                    width="15"
                    height="15"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={classNames('custom-dd-select__chevron', {
                        'custom-dd-select__chevron--open': isOpen,
                    })}>
                    <polyline points="6 9 12 15 18 9" />
                </svg>
            </button>

            {isOpen && (
                <div className="custom-dd-select__menu" role="listbox">
                    {options.map((opt) => {
                        const isSelected = opt.value === value;
                        return (
                            <button
                                key={opt.value}
                                type="button"
                                role="option"
                                aria-selected={isSelected}
                                className={classNames('custom-dd-select__option', {
                                    'custom-dd-select__option--active': isSelected,
                                })}
                                onClick={() => {
                                    setIsOpen(false);
                                    if (opt.value !== value) {
                                        onChange(opt.value);
                                    }
                                }}>
                                <span className="custom-dd-select__option-label">{opt.label}</span>
                                {isSelected && (
                                    <svg
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        strokeWidth="2.5"
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        className="custom-dd-select__check">
                                        <polyline points="20 6 9 17 4 12" />
                                    </svg>
                                )}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default CustomDropdownSelect;
