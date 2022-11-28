// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useState } from 'react';
import Button from '../button';
import Input from '../Input';

const DeleteItemsModal = ({ title, desc, handleDeleteSelected, buttontxt, textToConfirm }) => {
    const [confirm, setConfirm] = useState('');

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter' && confirm === (textToConfirm || 'permanently delete')) {
                handleDeleteSelected();
            }
        };
        document.addEventListener('keydown', keyDownHandler);
        return () => {
            document.removeEventListener('keydown', keyDownHandler);
        };
    }, [confirm]);

    return (
        <div className="delete-modal-wrapper">
            <p className="title">{title}</p>
            <p className="desc">{desc}</p>
            <div className="confirm-section">
                <p>
                    Please type <b>{textToConfirm || 'permanently delete'}</b> to confirm.
                </p>
                <Input
                    placeholder={textToConfirm || 'permanently delete'}
                    autoFocus={true}
                    type="text"
                    radiusType="semi-round"
                    colorType="black"
                    backgroundColorType="none"
                    borderColorType="gray-light"
                    height="48px"
                    onBlur={(e) => setConfirm(e.target.value)}
                    onChange={(e) => setConfirm(e.target.value)}
                    value={confirm}
                />
            </div>
            <div className="buttons">
                <Button
                    width="100%"
                    height="34px"
                    placeholder={buttontxt}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    disabled={confirm !== (textToConfirm || 'permanently delete')}
                    onClick={() => handleDeleteSelected()}
                />
            </div>
        </div>
    );
};

export default DeleteItemsModal;
