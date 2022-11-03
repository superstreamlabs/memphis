// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useState } from 'react';
import Button from '../button';
import Input from '../Input';

const DeleteItemsModal = ({ title, desc, handleDeleteSelected, buttontxt, textToConfirm }) => {
    const [confirm, setConfirm] = useState('');
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
