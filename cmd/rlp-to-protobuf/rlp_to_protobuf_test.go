package main

import (
	"errors"
	"flag"
	"fmt"
	"math/big"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

var _ = Describe("RLPtoProtobufCommand", func() {

	var (
		ctrl   *gomock.Controller
		src    *db.MockISubstateDB
		dst    *db.MockISubstateDB
		input0 *substate.Substate
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		src = db.NewMockISubstateDB(ctrl)
		dst = db.NewMockISubstateDB(ctrl)

		input0 = &substate.Substate{
			InputSubstate:  substate.NewWorldState(),
			OutputSubstate: substate.NewWorldState(),
			Env: &substate.Env{
				Coinbase:   types.Address{1},
				Difficulty: new(big.Int).SetUint64(1),
				GasLimit:   1,
				Number:     1,
				Timestamp:  1,
				BaseFee:    new(big.Int).SetUint64(1),
			},
			Message: substate.NewMessage(
				1,
				true,
				new(big.Int).SetUint64(1),
				1,
				types.Address{1},
				new(types.Address),
				new(big.Int).SetUint64(1),
				[]byte{1},
				nil,
				types.AccessList{},
				new(big.Int).SetUint64(1),
				new(big.Int).SetUint64(1),
				new(big.Int).SetUint64(1),
				make([]types.Hash, 0),
			),
			Result:      substate.NewResult(1, types.Bloom{}, []*types.Log{}, types.Address{}, 1),
			Block:       37_534_834,
			Transaction: 1,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("parsing fails", func() {
		It("should return an error", func() {
			set := flag.NewFlagSet("test", 0)
			_ = set.String(BlockSegmentFlag.Name, "0-abc", "")
			_ = set.String(WorkersFlag.Name, "1", "")
			ctx := cli.NewContext(&cli.App{}, set, nil)

			command := rlpToProtobufCommand{
				src: src,
				dst: dst,
				ctx: ctx,
			}

			err := command.execute()
			expected := errors.New("invalid block segment string: \"0-abc\"")

			Expect(err).To(Equal(expected))
		})
	})

	When("setting encoding fails", func() {
		It("should return an error", func() {
			set := flag.NewFlagSet("test", 0)
			_ = set.String(BlockSegmentFlag.Name, "0-2", "")
			_ = set.String(WorkersFlag.Name, "1", "")
			ctx := cli.NewContext(&cli.App{}, set, nil)
			mockErr := errors.New("error")

			dst.EXPECT().SetSubstateEncoding("protobuf").Return(nil, mockErr)

			command := rlpToProtobufCommand{
				src: src,
				dst: dst,
				ctx: ctx,
			}

			err := command.execute()

			Expect(err).To(Equal(mockErr))
		})
	})

	When("upgrading fails", func() {
		It("should return an error", func() {
			set := flag.NewFlagSet("test", 0)
			_ = set.String(BlockSegmentFlag.Name, "0-2", "")
			_ = set.String(WorkersFlag.Name, "1", "")
			ctx := cli.NewContext(&cli.App{}, set, nil)

			mockErr := errors.New("error")
			command := rlpToProtobufCommand{
				src: src,
				dst: dst,
				ctx: ctx,
			}

			dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.SubstateDB{}, nil)
			gomock.InOrder(
				src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil),
				src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil),
				src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil),
			)

			dst.EXPECT().PutSubstate(input0).Return(nil)
			dst.EXPECT().PutSubstate(input0).Return(nil)
			dst.EXPECT().PutSubstate(input0).Return(mockErr)
			err := command.execute()

			expected := fmt.Errorf("rlp-to-protobuf: 2_0: %w", fmt.Errorf("failed to put substate: %w", mockErr))
			Expect(err).To(Equal(expected))
		})
	})

	Context("with sequential processing", func() {

		When("all operations are successful", func() {
			It("should succeed", func() {
				set := flag.NewFlagSet("test", 0)
				_ = set.String(BlockSegmentFlag.Name, "0-2", "")
				_ = set.String(WorkersFlag.Name, "1", "")
				ctx := cli.NewContext(&cli.App{}, set, nil)

				command := rlpToProtobufCommand{
					src: src,
					dst: dst,
					ctx: ctx,
				}

				dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.SubstateDB{}, nil)
				gomock.InOrder(
					src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
						0: input0,
					}, nil),
					src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
						0: input0,
					}, nil),
					src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
						0: input0,
					}, nil),
				)

				dst.EXPECT().PutSubstate(input0).Return(nil).Times(3)

				err := command.execute()

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("error occurs", func() {
			It("should returns an error", func() {
				set := flag.NewFlagSet("test", 0)
				_ = set.String(BlockSegmentFlag.Name, "0-2", "")
				_ = set.String(WorkersFlag.Name, "1", "")
				ctx := cli.NewContext(&cli.App{}, set, nil)
				mockErr := errors.New("error")

				command := rlpToProtobufCommand{
					src: src,
					dst: dst,
					ctx: ctx,
				}

				dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.SubstateDB{}, nil)
				gomock.InOrder(
					src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
						0: input0,
					}, nil),
					src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, mockErr),
					src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
						0: input0,
					}, nil),
				)

				dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

				err := command.execute()
				Expect(err).To(Equal(mockErr))
			})
		})

	})

	Context("with parallel processing", func() {
		When("all operations are successful", func() {
			It("should succeed", func() {
				set := flag.NewFlagSet("test", 0)
				_ = set.String(BlockSegmentFlag.Name, "0-2", "")
				_ = set.String(WorkersFlag.Name, "4", "")
				ctx := cli.NewContext(&cli.App{}, set, nil)

				command := rlpToProtobufCommand{
					src: src,
					dst: dst,
					ctx: ctx,
				}

				dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.SubstateDB{}, nil)

				src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil)
				src.EXPECT().GetBlockSubstates(uint64(1)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil)
				src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil)

				dst.EXPECT().PutSubstate(input0).Return(nil).Times(3)

				err := command.execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("error occurs", func() {
			It("should returns an error", func() {
				set := flag.NewFlagSet("test", 0)
				_ = set.String(BlockSegmentFlag.Name, "0-2", "")
				_ = set.String(WorkersFlag.Name, "4", "")
				ctx := cli.NewContext(&cli.App{}, set, nil)
				mockErr := errors.New("error")

				command := rlpToProtobufCommand{
					src: src,
					dst: dst,
					ctx: ctx,
				}

				dst.EXPECT().SetSubstateEncoding("protobuf").Return(&db.SubstateDB{}, nil)

				src.EXPECT().GetBlockSubstates(uint64(0)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil)
				src.EXPECT().GetBlockSubstates(uint64(1)).Return(nil, mockErr)
				src.EXPECT().GetBlockSubstates(uint64(2)).Return(map[int]*substate.Substate{
					0: input0,
				}, nil)

				dst.EXPECT().PutSubstate(input0).Return(nil).Times(2)

				err := command.execute()
				Expect(err).To(Equal(mockErr))
			})
		})
	})

})
